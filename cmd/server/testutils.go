//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

// +test windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import "os"

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
	os.Remove(database)
	os.RemoveAll(storage)
	os.Remove(testConfig)
}
