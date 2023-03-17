//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/knoxite/knoxite/cmd/server/config"
	"github.com/knoxite/knoxite/cmd/server/utils"
	gap "github.com/muesli/go-app-paths"
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
	certsPath     string
	logPath       string
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

			if cfg.UseHTTPS {
				if err := initCerts(); err != nil {
					defer os.Remove(cfg.DBFileName)
					defer os.Remove(certsPath)
					ErrorLogger.Println(err.Error())
					return err
				}
			}
			return nil
		},
	}
)

func setPaths() error {
	if os.Getenv("APP_ENV") == "development" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		certsPath = filepath.Join(wd, "cmd", "server", "certs")
		logPath = filepath.Join(wd, "cmd", "server", "logs")
	} else if os.Getenv("APP_ENV") == "test" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		certsPath = filepath.Join(wd, "certs")
		logPath = filepath.Join(wd, "logs")
	} else {
		scope := gap.NewScope(gap.User, "knoxite-server")
		dataDirs, err := scope.DataDirs()
		if err != nil {
			return err
		}
		logPath, err = scope.LogPath("knoxite-server.log")
		if err != nil {
			return err
		}

		if _, err = os.Stat(filepath.Dir(logPath)); os.IsNotExist(err) {
			if err := os.Mkdir(filepath.Dir(logPath), 0755); err != nil {
				return err
			}
		}

		certsPath = filepath.Join(dataDirs[0], "certs")
		os.Setenv("APP_ENV", "production")
	}

	return nil
}

func init() {
	if err := setPaths(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	setupCmd.PersistentFlags().StringVarP(&cfg.AdminPassword, "password", "p", "", "Admin password")
	setupCmd.PersistentFlags().StringVarP(&cfg.AdminUserName, "username", "u", "", "Admin username")
	setupCmd.PersistentFlags().StringVarP(&cfg.DBFileName, "dbfilename", "d", "", "File name for sqlite database")
	setupCmd.PersistentFlags().StringVarP(&cfg.HTTPPort, "httpport", "P", "42024", "Port for server api")
	setupCmd.PersistentFlags().StringVarP(&cfg.HTTPSPort, "httpsport", "t", "42025", "Port for server api")
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

func InitLogger() {
	logFilePath := filepath.Join(logPath, "knoxite-server.log")
	if os.Getenv("APP_ENV") == "production" {
		logFilePath = logPath
	}
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

func initCerts() error {
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

	f, err := os.OpenFile(filepath.Join(certsPath, "knoxite-server-ca-cert.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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

	f, err = os.OpenFile(filepath.Join(certsPath, "knoxite-server-ca-key.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
	f, err = os.OpenFile(filepath.Join(certsPath, "knoxite-server-cert.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
	f, err = os.OpenFile(filepath.Join(certsPath, "knoxite-server-key.pem"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
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
