//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/knoxite/knoxite/cmd/server/config"
	"github.com/knoxite/knoxite/cmd/server/utils"
	"github.com/natefinch/lumberjack"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PackageJSON struct {
	Name         string      `json:"name"`
	Version      string      `json:"version"`
	Private      bool        `json:"private"`
	Dependencies interface{} `json:"dependencies"`
	Scripts      interface{} `json:"scripts"`
	Proxy        string      `json:"proxy"`
	ESLintConfig interface{} `json:"eslintConfig"`
	BrowsersList interface{} `json:"browserslist"`
}

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
	wd            string
	setupCmd      = &cobra.Command{
		Use: "setup",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitLogger()
			if err := initHostname(); err != nil {
				ErrorLogger.Println(err.Error())
				return err
			}

			if err := cfg.Save(configURL); err != nil {
				ErrorLogger.Println(err.Error())
				return err
			}

			if err := initDB(cfg.DBFileName); err != nil {
				ErrorLogger.Println(err.Error())
				return err
			}

			if err := initStoragePath(cfg.StoragesPath); err != nil {
				defer os.Remove(cfg.DBFileName)
				ErrorLogger.Println(err.Error())
				return err
			}

			if os.Getenv("APP_ENV") == "production" {
				if err := initYarnInstallAndBuild(); err != nil {
					defer os.Remove(cfg.DBFileName)
					defer os.Remove(filepath.Join(wd, "cmd", "server", "ui", "build"))
					ErrorLogger.Println(err.Error())
					return err
				}
			}

			if err := initDotEnv(); err != nil {
				defer os.Remove(cfg.DBFileName)
				defer os.Remove(filepath.Join(wd, "cmd", "server", "ui", "build"))
				ErrorLogger.Println(err.Error())
				return err
			}

			if cfg.UseHTTPS {
				certsPath := filepath.Join(wd, "certs")
				if err := initCerts(certsPath); err != nil {
					defer os.Remove(cfg.DBFileName)
					defer os.Remove(filepath.Join(wd, "cmd", "server", "ui", "build"))
					defer os.Remove(certsPath)
					ErrorLogger.Println(err.Error())
					return err
				}
			}
			return nil
		},
	}
)

func init() {
	wd, _ = os.Getwd()

	setupCmd.PersistentFlags().StringVarP(&cfg.AdminPassword, "password", "p", "", "Admin password")
	setupCmd.PersistentFlags().StringVarP(&cfg.AdminUserName, "username", "u", "", "Admin username")
	setupCmd.PersistentFlags().StringVarP(&cfg.DBFileName, "dbfilename", "d", "", "File name for sqlite database")
	setupCmd.PersistentFlags().StringVarP(&cfg.AdminUIPort, "port", "P", "42024", "Port for server api")
	setupCmd.PersistentFlags().StringVarP(&cfg.StoragesPath, "storagepath", "s", "", "Path to client storages")
	setupCmd.PersistentFlags().StringVarP(&configURL, "configurl", "C", config.DefaultPath(), "Path to configuration file")
	setupCmd.PersistentFlags().BoolVarP(&cfg.UseHostname, "usehostname", "n", true, "Use hostname and dns, default=true")
	setupCmd.PersistentFlags().BoolVarP(&cfg.UseHTTPS, "usehttps", "S", true, "Use https, default=true")

	setupCmd.MarkFlagRequired("password")
	setupCmd.MarkFlagRequired("username")
	setupCmd.MarkFlagRequired("dbfilename")
	setupCmd.MarkFlagRequired("port")
	setupCmd.MarkFlagRequired("repopath")

	RootCmd.AddCommand(setupCmd)
}

func initDB(dbURL string) error {
	db, err := gorm.Open(sqlite.Open(dbURL), &gorm.Config{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&Client{})
	if err != nil {
		return err
	}
	return nil
}

func initHostname() error {
	if cfg.UseHostname {
		hostname, err := os.Hostname()
		if err != nil {
			return err
		}
		cfg.AdminHostname = hostname
	}
	return nil
}

// Note: the repoURL, since it is a folder, needs to end with a "/", e.g. /tmp/repositories/
func initStoragePath(storageURL string) error {
	path, err := utils.PathToUrl(storageURL)
	if err != nil {
		return err
	}

	storageDir := filepath.Join(path.Path)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return err
	}

	return nil
}

func initYarnInstallAndBuild() error {
	uiFilePath := filepath.Join(wd, "ui")
	InfoLogger.Println("Installing yarn packages (this could take a while) ...")
	cmd := exec.Command("yarn", "install")
	cmd.Dir = uiFilePath
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return err
	}
	scheme := "http://"
	if cfg.UseHTTPS {
		scheme = "https://"
	}
	httpProxy := scheme + "localhost:" + cfg.AdminUIPort

	// Rebuild package.json
	file, err := os.ReadFile(filepath.Join(uiFilePath, "package.json")) //Read File
	if err != nil {
		return err
	}
	var p PackageJSON
	json.Unmarshal(file, &p)
	p.Proxy = httpProxy
	result, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(uiFilePath, "package.json"), result, 0644)
	if err != nil {
		return err
	}

	InfoLogger.Println("Building admin user interface (this could take a while) ...")
	cmd = exec.Command("yarn", "build")
	cmd.Dir = uiFilePath
	cmd.Stdout = os.Stdout
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func InitLogger() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	logFilePath := filepath.Join(wd, "cmd", "server", "logs", "log.txt")
	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	log.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
}

func initCerts(certsPath string) error {
	if _, err := os.Stat(certsPath); os.IsNotExist(err) {
		err := os.Mkdir(certsPath, 0755)
		if err != nil {
			return err
		}
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return err
	}
	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"Knoxite"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", cfg.AdminHostname},
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	f, err := os.OpenFile(filepath.Join(certsPath, "ca-cert.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(caPEM.Bytes())
	if err != nil {
		return err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	f, err = os.OpenFile(filepath.Join(certsPath, "ca-key.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(caPrivKeyPEM.Bytes())
	if err != nil {
		return err
	}

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:  []string{"Knoxite"},
			Country:       []string{"US"},
			Province:      []string{""},
			Locality:      []string{"San Francisco"},
			StreetAddress: []string{"Golden Gate Bridge"},
			PostalCode:    []string{"94016"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		DNSNames:     []string{"localhost", cfg.AdminHostname},
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	f, err = os.OpenFile(filepath.Join(certsPath, "cert.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(certPEM.Bytes())
	if err != nil {
		return err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	f, err = os.OpenFile(filepath.Join(certsPath, "key.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(certPrivKeyPEM.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func initDotEnv() error {
	f, err := os.OpenFile(filepath.Join(wd, "cmd", "server", "ui", ".env"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	scheme := "http"
	if cfg.UseHTTPS {
		scheme = "https"
	}

	_, err = f.WriteString(
		"ADMIN_HOSTNAME=" + cfg.AdminHostname + "\n" +
			"ADMIN_UI_PORT=" + cfg.AdminUIPort + "\n" +
			"SERVER_SCHEME=" + scheme + "\n",
	)
	if err != nil {
		return err
	}
	return nil
}
