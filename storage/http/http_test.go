//go:build backend
// +build backend

/*
 * knoxite
 *     Copyright (c) 2016-2022, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2021-2022, Raschaad Yassine <Raschaad@gmx.de>
 *
 *   For license see LICENSE
 */
package http

import (
	"os"
	"testing"

	"github.com/knoxite/knoxite/storage"
)

func TestMain(m *testing.M) {
	knoxiteurl := os.Getenv("KNOXITE_HTTP_URL")
	if len(knoxiteurl) == 0 {
		panic("no backend configured")
	}

	backendTest = &storage.BackendTest{
		URL:         knoxiteurl,
		Protocols:   []string{"http", "https"},
		Description: "knoxite server storage",
		TearDown: func(tb *storage.BackendTest) {
			db := tb.Backend.(*HTTPStorage)
			err := db.DeleteFile(db.Path)
			if err != nil {
				panic(err)
			}
		},
	}

	storage.RunBackendTester(backendTest, m)
}

var (
	backendTest *storage.BackendTest
)

func TestStorageNewBackend(t *testing.T) {
	backendTest.NewBackendTest(t)
}

func TestStorageLocation(t *testing.T) {
	backendTest.LocationTest(t)
}

func TestStorageProtocols(t *testing.T) {
	backendTest.ProtocolsTest(t)
}

func TestStorageDescription(t *testing.T) {
	backendTest.DescriptionTest(t)
}

func TestStorageInitRepository(t *testing.T) {
	backendTest.InitRepositoryTest(t)
}

func TestStorageSaveRepository(t *testing.T) {
	backendTest.SaveRepositoryTest(t)
}

func TestAvailableSpace(t *testing.T) {
	backendTest.AvailableSpaceTest(t)
}

func TestStorageSaveSnapshot(t *testing.T) {
	backendTest.SaveSnapshotTest(t)
}

func TestStorageStoreChunk(t *testing.T) {
	backendTest.StoreChunkTest(t)
}

func TestStorageDeleteChunk(t *testing.T) {
	backendTest.DeleteChunkTest(t)
}
