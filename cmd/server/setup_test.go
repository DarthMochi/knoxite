//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"os"
	"testing"
)

var (
	testUsername = "abc"
	testPassword = "123"
	testDatabase = "test.db"
	testStorage  = "testdata/repositories/"
	testPort     = "8080"
	testConfig   = "testdata/knoxite-server.config"
)

func Test_ExecuteCommand(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	_, err = os.Stat(testDatabase)
	if err != nil {
		t.Errorf("expected error to be nil, database file missing, got %v", err)
	}
}
