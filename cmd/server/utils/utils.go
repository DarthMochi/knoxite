package utils

import (
	"net/url"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/bcrypt"
)

func Exist(file string) bool {
	_, err := os.Stat(file)
	return err == nil
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
