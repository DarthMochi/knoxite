//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

/*
 * knoxite
 *     Copyright (c) 2016-2022, Christian Muehlhaeuser <muesli@gmail.com>
 *     Copyright (c) 2021-2022, Raschaad Yassine <Raschaad@gmx.de>
 *
 *   For license see LICENSE
 */

package http

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/knoxite/knoxite"
	gap "github.com/muesli/go-app-paths"
)

// HTTPStorage stores data on a remote HTTP server.
type HTTPStorage struct {
	url url.URL
	knoxite.StorageFilesystem
}

type BackendClient struct {
	Name      string
	AuthCode  string
	Quota     uint64
	UsedSpace uint64
}

func init() {
	knoxite.RegisterStorageBackend(&HTTPStorage{})
}

// NewBackend returns a HTTPStorage backend.
func (*HTTPStorage) NewBackend(u url.URL) (knoxite.Backend, error) {
	authCode := u.User.Username()
	if authCode == "" {
		return &HTTPStorage{}, knoxite.ErrInvalidUsername
	}

	backend := HTTPStorage{url: u}

	fs, err := knoxite.NewStorageFilesystem("/", &backend)
	if err != nil {
		return &HTTPStorage{}, err
	}
	backend.StorageFilesystem = fs

	return &backend, nil
}

// Location returns the type and location of the repository.
func (backend *HTTPStorage) Location() string {
	return backend.url.String()
}

// Close the backend.
func (backend *HTTPStorage) Close() error {
	return nil
}

// Protocols returns the Protocol Schemes supported by this backend.
func (backend *HTTPStorage) Protocols() []string {
	return []string{"http", "https"}
}

// Description returns a user-friendly description for this backend.
func (backend *HTTPStorage) Description() string {
	return "knoxite server storage"
}

func (backend *HTTPStorage) getHTTPClient() *http.Client {
	if backend.url.Scheme == "https" {
		client := backend.getHTTPSClient()
		if client != nil {
			return client
		}
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}
	return client
}

func (backend *HTTPStorage) getHTTPSClient() *http.Client {
	var certDirPath string

	scope := gap.NewScope(gap.User, "knoxite")
	dataPaths, err := scope.DataDirs()
	if err != nil {
		return nil
	}

	if len(dataPaths) > 0 {
		certDirPath = dataPaths[0]
	}

	if _, err = os.Stat(certDirPath); os.IsNotExist(err) {
		if err = os.MkdirAll(certDirPath, 0755); err != nil {
			return nil
		}
	}

	certPath := filepath.Join(certDirPath, "knoxite-server-cert.pem")

	var cert []byte
	tlsConfig := &tls.Config{
		RootCAs:            x509.NewCertPool(),
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}

	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		url := strings.Replace(backend.url.String(), "https", "http", 1)
		req, err := http.NewRequest(http.MethodGet, url+"/download_cert", nil)
		if err != nil {
			return nil
		}
		req.Header.Set("Authorization", "Bearer "+backend.url.User.Username())
		httpClient := &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := httpClient.Do(req)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusOK {
			return nil
		}
		defer resp.Body.Close()

		cert, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil
		}

		if err = ioutil.WriteFile(certPath, cert, 0600); err != nil {
			return nil
		}
	} else {
		cert, err = ioutil.ReadFile(certPath)
		if err != nil {
			return nil
		}
	}

	if cert != nil {
		ok := tlsConfig.RootCAs.AppendCertsFromPEM(cert)
		if !ok {
			return nil
		}

		return client
	}

	return nil
}

// AvailableSpace returns the free space on this backend.
func (backend *HTTPStorage) AvailableSpace() (uint64, error) {
	client, err := backend.GetClientInfo()
	if err != nil {
		return 0, err
	}

	return client.Quota - client.UsedSpace, nil
}

func (backend *HTTPStorage) CreatePath(path string) error {
	var httpClient = backend.getHTTPClient()

	req, err := http.NewRequest(http.MethodGet, backend.url.String()+"/mkdir"+path, nil)
	if err != nil {
		return knoxite.ErrInvalidRepositoryURL
	}
	req.Header.Set("Authorization", "Bearer "+backend.url.User.Username())
	_, err = httpClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (backend *HTTPStorage) Stat(path string) (uint64, error) {
	var httpClient = backend.getHTTPClient()

	req, err := http.NewRequest(http.MethodGet, backend.url.String()+"/stat/"+path, nil)
	if err != nil {
		return 0, knoxite.ErrInvalidRepositoryURL
	}
	req.Header.Set("Authorization", "Bearer "+backend.url.User.Username())
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var file struct {
		Path string
		Size int64
	}
	err = json.NewDecoder(resp.Body).Decode(&file)
	if err != nil {
		return 0, err
	}

	return uint64(file.Size), nil
}

func (backend *HTTPStorage) ReadFile(path string) ([]byte, error) {
	var httpClient = backend.getHTTPClient()

	req, err := http.NewRequest(http.MethodGet, backend.url.String()+"/download/"+path, nil)
	if err != nil {
		return nil, knoxite.ErrInvalidRepositoryURL
	}
	req.Header.Set("Authorization", "Bearer "+backend.url.User.Username())
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("file not found")
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (backend *HTTPStorage) WriteFile(path string, data []byte) (size uint64, err error) {
	var httpClient = backend.getHTTPClient()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	fw, err := writer.CreateFormField("uploadfile")
	if err != nil {
		return 0, err
	}

	_, err = fw.Write(data)
	if err != nil {
		return 0, err
	}
	writer.Close()

	request, err := http.NewRequest(http.MethodPost, backend.url.String()+"/upload", body)
	if err != nil {
		return 0, err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set("Path", path)
	request.Header.Set("Authorization", "Bearer "+backend.url.User.Username())

	response, err := httpClient.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	var file struct {
		Path string
		Size int64
	}
	json.NewDecoder(response.Body).Decode(&file)

	return uint64(file.Size), nil
}

func (backend *HTTPStorage) DeleteFile(path string) error {
	var httpClient = backend.getHTTPClient()

	req, err := http.NewRequest(http.MethodDelete, backend.url.String()+"/delete", nil)
	if err != nil {
		return knoxite.ErrInvalidRepositoryURL
	}
	req.Header.Set("Authorization", "Bearer "+backend.url.User.Username())
	req.Header.Set("Path", path)
	_, err = httpClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (backend *HTTPStorage) GetClientInfo() (BackendClient, error) {
	var httpClient = backend.getHTTPClient()

	req, err := http.NewRequest(http.MethodGet, backend.url.String()+"/getClientByAuthCode", nil)
	if err != nil {
		return BackendClient{}, knoxite.ErrInvalidRepositoryURL
	}
	req.Header.Set("Authorization", "Bearer "+backend.url.User.Username())
	resp, err := httpClient.Do(req)
	if err != nil {
		return BackendClient{}, err
	}
	defer resp.Body.Close()

	var client BackendClient
	err = json.NewDecoder(resp.Body).Decode(&client)
	if err != nil {
		return BackendClient{}, err
	}
	return client, nil
}

// LoadChunkIndex reads the chunk-index.
func (backend *HTTPStorage) LoadChunkIndex() ([]byte, error) {

	var httpClient = backend.getHTTPClient()

	req, err := http.NewRequest(http.MethodGet, backend.url.String()+"/download/chunks/index", nil)
	if err != nil {
		return []byte{}, knoxite.ErrInvalidRepositoryURL
	}
	req.Header.Set("Authorization", "Bearer "+backend.url.User.Username())
	resp, err := httpClient.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return []byte{}, fmt.Errorf("no chunk index created yet")
	}

	return io.ReadAll(resp.Body)
}

// SaveChunkIndex stores the chunk-index.
func (backend *HTTPStorage) SaveChunkIndex(data []byte) error {
	_, err := backend.WriteFile("/chunks/index", data)
	if err != nil {
		return knoxite.ErrStoreChunkFailed
	}
	return nil
}
