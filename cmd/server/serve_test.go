//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/knoxite/knoxite/cmd/server/utils"
)

var (
	app       = &App{}
	newClient = &Client{
		Name: "Testclient",
	}
)

func TestErrors(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)

	body := url.Values{
		"name":  {newClient.Name},
		"quota": []string{"1000000000"},
	}

	request := httptest.NewRequest(http.MethodPost, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + "1:" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()
	app.createClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusForbidden {
		t.Errorf("Want status '%d', got '%d'", http.StatusForbidden, responseRecorder.Result().StatusCode)
	}

	body = url.Values{}

	request = httptest.NewRequest(http.MethodPost, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	baseAuthEnc = b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	request.Body = nil
	app.createClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("Want status '%d', got '%d'", http.StatusInternalServerError, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"name":  {newClient.Name},
		"quota": []string{"100000000000000000000000"},
	}

	request = httptest.NewRequest(http.MethodPost, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.createClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"name":  {newClient.Name},
		"quota": []string{"1000000000000000"},
	}

	request = httptest.NewRequest(http.MethodPost, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.createClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"name":  {newClient.Name + "/.."},
		"quota": []string{"1000000000"},
	}

	request = httptest.NewRequest(http.MethodPost, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.createClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Result().StatusCode)
	}

	createClient(t)

	request = httptest.NewRequest(http.MethodGet, "/api/clients", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic h"+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.getAllClients(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusForbidden {
		t.Errorf("Want status '%d', got '%d'", http.StatusForbidden, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"id": []string{"1"},
	}

	request = httptest.NewRequest(http.MethodGet, "/api/clients/1", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic h"+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.getClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusForbidden {
		t.Errorf("Want status '%d', got '%d'", http.StatusForbidden, responseRecorder.Result().StatusCode)
	}

	request = httptest.NewRequest(http.MethodGet, "/api/clients", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic h"+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.getClientByAuthCode(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusForbidden {
		t.Errorf("Want status '%d', got '%d'", http.StatusForbidden, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"name":  {newClient.Name},
		"quota": []string{"1000000000"},
	}

	request = httptest.NewRequest(http.MethodPut, "/api/clients/1", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic h"+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.updateClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusForbidden {
		t.Errorf("Want status '%d', got '%d'", http.StatusForbidden, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"name":  {newClient.Name},
		"quota": []string{"10000000000000000000000000000000"},
	}

	request = httptest.NewRequest(http.MethodPut, "/api/clients/1", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.updateClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"name":  {newClient.Name},
		"quota": []string{"10000000000000000000"},
	}

	request = httptest.NewRequest(http.MethodPut, "/api/clients/1", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.updateClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"name":  {newClient.Name + "/.."},
		"quota": []string{"1000000000"},
	}

	request = httptest.NewRequest(http.MethodPut, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.updateClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Want status '%d', got '%d'", http.StatusBadRequest, responseRecorder.Result().StatusCode)
	}

	body = url.Values{
		"id": []string{"1"},
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/clients/1", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic h"+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.deleteClient(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusForbidden {
		t.Errorf("Want status '%d', got '%d'", http.StatusForbidden, responseRecorder.Result().StatusCode)
	}

	request = httptest.NewRequest(http.MethodDelete, "/api/login", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Authorization", "Basic h"+baseAuthEnc)
	responseRecorder = httptest.NewRecorder()
	app.login(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("Want status '%d', got '%d'", http.StatusUnauthorized, responseRecorder.Result().StatusCode)
	}
}

func TestCreateClient(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
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

		u, err := url.Parse(fmt.Sprintf("/api/clients/%d", client.ID))
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
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
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
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
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
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
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
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
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
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
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
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
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

func TestGetAllClients(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/clients", nil)
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()
	app.getAllClients(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
	}

	var jsonBody []Client
	err = json.NewDecoder(responseRecorder.Body).Decode(&jsonBody)
	if err != nil {
		t.Errorf("unexpected error when parsing body to json, failed with: %v", err)
	}

	if len(jsonBody) == 0 {
		t.Errorf("clients are empty, there should be at least one")
	}
}

func TestGetClient(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)

	body := map[string]string{
		"id": "1",
	}

	request := httptest.NewRequest(http.MethodGet, "/api/clients", nil)
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()
	request = mux.SetURLVars(request, body)
	app.getClient(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
	}

	var jsonClient Client
	err = json.NewDecoder(responseRecorder.Body).Decode(&jsonClient)
	if err != nil {
		t.Errorf("unexpected error when parsing body to json, failed with: %v", err)
	}

	if jsonClient.ID != 1 {
		t.Errorf("got wrong client")
	}
}

func TestGetClientByAuthCode(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)

	request := httptest.NewRequest(http.MethodGet, "/getClientByAuthCode", nil)
	request.Header.Set("Authorization", "Bearer "+newClient.AuthCode)
	responseRecorder := httptest.NewRecorder()
	app.getClientByAuthCode(responseRecorder, request)
	if responseRecorder.Result().StatusCode != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
	}

	bodyBytes, err := io.ReadAll(responseRecorder.Body)
	if err != nil {
		t.Errorf("unexpected error when reading body, failed with: %v", err)
	}

	var jsonClient Client
	err = json.Unmarshal(bodyBytes, &jsonClient)
	if err != nil {
		t.Errorf("unexpected error when parsing body to json, failed with: %v", err)
	}

	if jsonClient.ID != 1 {
		t.Errorf("got wrong client")
	}
}

func TestUpdateClient(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)

	body := url.Values{
		"name":  []string{"another_testname"},
		"quota": []string{"100000000"},
	}

	request := httptest.NewRequest(http.MethodPut, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "1",
	}
	request = mux.SetURLVars(request, vars)
	app.updateClient(responseRecorder, request)
	var clients []Client
	app.DB.Find(&clients)

	updatedClient := &clients[len(clients)-1]

	if responseRecorder.Result().StatusCode != http.StatusNoContent {
		t.Errorf("Want status '%d', got '%d'", http.StatusNoContent, responseRecorder.Code)
	}

	if uint64(updatedClient.Quota) != uint64(100000000) {
		t.Errorf("Failed to update quota, expected '%d', got '%d'", 100000000, newClient.Quota)
	}

	if updatedClient.Name != "another_testname" {
		t.Errorf("Failed to update name, expected '%s', got '%s'", "another_testname", newClient.Name)
	}
}

func TestDeleteClient(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/clients/1", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()

	vars := map[string]string{
		"id": "1",
	}
	request = mux.SetURLVars(request, vars)
	app.deleteClient(responseRecorder, request)
	var clients []Client
	app.DB.Find(&clients)

	if responseRecorder.Result().StatusCode != http.StatusNoContent {
		t.Errorf("Want status '%d', got '%d'", http.StatusNoContent, responseRecorder.Code)
	}

	if len(clients) != 0 {
		t.Errorf("Client wasn't removed, expected '%d', got '%d'", 0, len(clients))
	}
}

func TestLogin(t *testing.T) {
	err := SetupServer(testUsername, testPassword, testDatabase, testStorage, testPort, testConfig)
	defer Cleanup(testDatabase, testStorage, testConfig)
	if err != nil {
		t.Errorf("expected error to be nil, got %v", err)
	}
	app.initialize(testDatabase)
	createClient(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/login", nil)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()

	app.login(responseRecorder, request)

	if responseRecorder.Result().StatusCode != http.StatusOK {
		t.Errorf("Want status '%d', got '%d'", http.StatusOK, responseRecorder.Code)
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
	bytes, err := io.ReadAll(file)
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

	request := httptest.NewRequest(http.MethodPost, "/api/clients", strings.NewReader(body.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hash, _ := utils.HashPassword(testPassword)
	baseAuthEnc := b64.StdEncoding.EncodeToString([]byte(testUsername + ":" + hash))
	request.Header.Add("Authorization", "Basic "+baseAuthEnc)
	responseRecorder := httptest.NewRecorder()
	app.createClient(responseRecorder, request)

	var clients []Client
	app.DB.Find(&clients)

	if len(clients) > 0 {
		newClient = &clients[len(clients)-1]
	}

	return *responseRecorder
}
