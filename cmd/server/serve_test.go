package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strings"
	"testing"
)

var (
	/*username   = "abc"
	password   = "123"
	database   = "test.db"
	repo       = "testdata/repositories/"
	port       = "8080"
	testConfig = "testdata/knoxite-server.config"*/
	app       = &App{}
	newClient = &Client{
		Name: "Testclient",
	}
)

func TestCreateClient(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize("test.db")

	responseRecorder := createClient(t)
	var clients []Client

	app.DB.Find(&clients)

	if len(clients) < 2 {
		t.Errorf("Client was not created")
	} else {
		client := clients[len(clients)-1]

		if client.Name != newClient.Name {
			t.Errorf("expected client name '%s', got '%s'", newClient.Name, client.Name)
		}

		if client.AuthCode == "" {
			t.Errorf("expected client authcode, got nothing")
		}

		location, err := responseRecorder.Result().Location()
		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}

		u, err := url.Parse(fmt.Sprintf("/clients/%d", client.ID))
		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}

		if location.Path != u.Path {
			t.Errorf("Location wanted '%s', got '%s'", u, location)
		}

		if responseRecorder.Code != http.StatusCreated {
			t.Errorf("Want status '%d', got '%d", http.StatusCreated, responseRecorder.Code)
		}
	}
}

func TestUpload(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	app.initialize(testDatabase)
	createClient(t)

	responseRecorder := uploadTestFileRequest(t, testDatabase)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Want status '%d', got '%d", http.StatusOK, responseRecorder.Code)
	}
}

func TestDownload(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	// defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	testfile := "test.db"
	createClient(t)
	uploadTestFileRequest(t, testfile)

	request := httptest.NewRequest(http.MethodGet, "/download/", strings.NewReader(testfile))
	responseRecorder := httptest.NewRecorder()
	app.download(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Want status '%d', got '%d", http.StatusOK, responseRecorder.Code)
	}
}

func uploadTestFileRequest(t *testing.T, testfile string) httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormFile("uploadfile", testfile)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	file, err := os.Open(path.Join(".", testfile))
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body.Bytes()))
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	responseRecorder := httptest.NewRecorder()
	app.upload(responseRecorder, request)

	return *responseRecorder
}

func createClient(t *testing.T) httptest.ResponseRecorder {
	body := url.Values{
		"name": {newClient.Name},
	}

	request := httptest.NewRequest(http.MethodPost, "/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	responseRecorder := httptest.NewRecorder()
	app.createClient(responseRecorder, request)
	var clients []Client
	app.DB.Find(&clients)

	newClient = &clients[len(clients)-1]

	return *responseRecorder
}
