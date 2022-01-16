//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/knoxite/knoxite/cmd/server/utils"
)

var (
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
	app.initialize(testDatabase)

	responseRecorder := createClient(t)
	var clients []Client

	app.DB.Find(&clients)

	if len(clients) < 1 {
		t.Errorf("Client was not created")
	} else {
		client := clients[len(clients)-1]

		if client.Name != newClient.Name {
			t.Errorf("expected client name '%s', got '%s'", newClient.Name, client.Name)
		}

		if client.AuthCode == "" {
			t.Errorf("expected client authcode, got nothing")
		}

		location := responseRecorder.Header().Get("Location")

		u, err := url.Parse(fmt.Sprintf("/clients/%d", client.ID))
		if err != nil {
			t.Errorf("expected error to be nil, got %v", err)
		}

		if location != u.Path {
			t.Errorf("Location wanted '%s', got '%s'", u, location)
		}

		// TODO: should be http.StatusCreated
		if responseRecorder.Code != http.StatusOK {
			t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Result().StatusCode)
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

	responseRecorder := uploadTestFileRequest(t, "loremipsum")

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
	}
}

func TestDownload(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	testfile := "loremipsum"
	createClient(t)
	uploadTestFileRequest(t, testfile)

	request := httptest.NewRequest(http.MethodGet, "/download/chunks/"+testfile, nil)
	responseRecorder := httptest.NewRecorder()
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	app.download(responseRecorder, request)

	if responseRecorder.Code != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
	}
}

func TestStat(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)
	testfile := "loremipsum"
	uploadTestFileRequest(t, testfile)

	request := httptest.NewRequest(http.MethodGet, "/size/chunks/", strings.NewReader(testfile))
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	responseRecorder := httptest.NewRecorder()
	app.getFileStats(responseRecorder, request)

	var file struct {
		Path string
		Size int64
	}
	json.NewDecoder(responseRecorder.Result().Body).Decode(&file)

	if file.Size < 1 {
		t.Errorf("Want size bigger than 1, got '%d'", file.Size)
	}
}

func TestMkdir(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)

	request := httptest.NewRequest(http.MethodGet, "/mkdir/chunks/1234", nil)
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	responseRecorder := httptest.NewRecorder()
	app.mkdir(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusCreated {
		t.Errorf("Want status '%d', got '%d'", http.StatusCreated, responseRecorder.Code)
	}
}

func TestDeleteFile(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)
	testfile := "loremipsum"
	uploadTestFileRequest(t, testfile)

	request := httptest.NewRequest(http.MethodDelete, "/delete", nil)
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	request.Header.Set("Path", "/chunks/"+testfile)
	responseRecorder := httptest.NewRecorder()
	app.delete(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
	}
}

func TestFilePathTraversal(t *testing.T) {
	err := setupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)
	uploadTestFileRequest(t, "loremipsum")

	request := httptest.NewRequest(http.MethodDelete, "/delete", nil)
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	request.Header.Set("Path", "/chunks/../../../../../../../../../../../chunks/loremipsum")
	responseRecorder := httptest.NewRecorder()
	app.delete(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/mkdir/chunks/../../../../../../../../../../../chunks/loremipsum", nil)
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	responseRecorder = httptest.NewRecorder()
	app.mkdir(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/size/chunks/../../../../../../../../../../../chunks/loremipsum", nil)
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	responseRecorder = httptest.NewRecorder()
	app.getFileStats(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Code)
	}

	request = httptest.NewRequest(http.MethodGet, "/download/chunks/../../../../../../../../../../../chunks/loremipsum", nil)
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	responseRecorder = httptest.NewRecorder()
	app.download(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Want status '%d', got '%d'", http.StatusNotFound, responseRecorder.Code)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	testfile := "loremipsum"

	fw, err := writer.CreateFormFile("uploadfile", testfile)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	file, err := os.Open(filepath.Join(".", "testdata", testfile))
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	_, err = io.Copy(fw, file)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	writer.Close()
	request = httptest.NewRequest(http.MethodPost, "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	request.Header.Set("Path", "/chunks/../../../../../../../../../../../chunks/"+testfile)
	responseRecorder = httptest.NewRecorder()
	app.upload(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Code)
	}
}

func uploadTestFileRequest(t *testing.T, testfile string) httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormField("uploadfile")
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}

	file, err := os.Open(filepath.Join(".", "testdata", testfile))
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	_, err = fw.Write(bytes)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	writer.Close()

	request := httptest.NewRequest(http.MethodPost, "/upload", body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	request.Header.Set("Path", "chunks/"+testfile)
	responseRecorder := httptest.NewRecorder()
	app.upload(responseRecorder, request)

	return *responseRecorder
}

func createClient(t *testing.T) httptest.ResponseRecorder {
	body := url.Values{
		"name":  {newClient.Name},
		"quota": []string{"1000000000"},
	}

	request := httptest.NewRequest(http.MethodPost, "/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()
	app.createClient(responseRecorder, request)
	var clients []Client
	app.DB.Find(&clients)

	newClient = &clients[len(clients)-1]

	return *responseRecorder
}
