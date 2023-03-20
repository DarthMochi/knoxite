//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

// +test windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func SetupServer(username string, password string, database string, storage string, port string, testConfig string) error {
	RootCmd.SetArgs([]string{
		"setup",
		"-u", username,
		"-p", password,
		"-d", database,
		"-s", storage,
		"-P", port,
		"-C", testConfig,
	})
	err := RootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func Cleanup(database string, storage string, testConfig string) {
	db, _ := gorm.Open(sqlite.Open(database))
	db.Migrator().DropTable(&Client{})
	os.RemoveAll(storage)
	os.Remove(testConfig)
}
