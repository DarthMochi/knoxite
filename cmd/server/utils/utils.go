//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/Luzifer/go-openssl/v4"
	"github.com/mitchellh/go-homedir"
	gap "github.com/muesli/go-app-paths"
	"golang.org/x/crypto/bcrypt"
)

func DefaultPath(appName, filename string) string {
	userScope := gap.NewScope(gap.User, appName)
	path, err := userScope.ConfigPath(filename)
	if err != nil {
		return filename
	}

	return path
}

func Exist(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

func ParseClientURL(clientId uint) (*url.URL, error) {
	return url.Parse(fmt.Sprintf("/clients/%d", clientId))
}

func LogRequest(r *http.Request) string {
	logString := r.Method
	logString += " " + r.URL.String()
	logString += " from " + r.RemoteAddr
	return logString
}

func PathToUrl(u string) (*url.URL, error) {
	url := &url.URL{}
	// Check if the given string starts with a protocol scheme. Prepend the file
	// scheme in case none is provided
	if !isUrl(u) {
		url.Scheme = "file"
		url.Path = u
	} else {
		// u = url.QueryEscape(u)
		var err error
		url, err = url.Parse(u)
		if err != nil {
			return url, err
		}
	}

	// In case some other path elements have wrongfully been interpreted as Host
	// part of the url
	if url.Host != "" {
		url.Path = url.Host + url.Path
		url.Host = ""
	}

	// Expand tilde to the users home directory
	// This is needed in case the shell is unable to expand the path to the users
	// home directory for inputs like these:
	// crypto://password@~/path/to/config
	var err error
	url.Path, err = homedir.Expand(url.Path)
	if err != nil {
		return nil, err
	}
	return url, nil
}

func isUrl(str string) bool {
	if _, err := url.Parse(str); err != nil {
		return false
	}

	return strings.Contains(str, "://")
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func ByteArrDiff(original, new []byte) int64 {
	// fmt.Println("new chunk size: ", len(new))
	// fmt.Println("old chunk size: ", len(original))
	// Checks, if new byte array is initialized.
	if (new == nil) || len(new) < 1 {
		return int64(len(new))
	}

	// Checks, if original byte array is empty and returns the length of the new byte array, if original is empty.
	if (original == nil) || len(original) < 1 {
		return int64(len(new))
	}

	smallerArr := original
	biggerArr := new
	lenO := len(original)
	lenN := len(new)

	if lenN < lenO {
		biggerArr = original
		smallerArr = new
	}

	// Calculates the difference between the two.
	var diff int64 = 0
	for i := range smallerArr {
		if int16(smallerArr[i]) != int16(biggerArr[i]) {
			diff += 1
		}
	}

	diff -= int64(lenO) - int64(lenN)

	return diff
}

func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}

	return x
}

// GetLocalIP returns the non loopback local IP of the host.
func GetLocalIP() net.IP {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP
			}
		}
	}
	return nil
}

func DecryptAES(key string, ct string) (string, error) {
	o := openssl.New()

	dec, err := o.DecryptBytes(key, []byte(ct), openssl.BytesToKeyMD5)
	if err != nil {
		return "", err
	}
	return string(dec), nil
}

func GenerateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
