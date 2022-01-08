//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoxite/knoxite/cmd/server/config"
	"github.com/knoxite/knoxite/cmd/server/utils"
	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type App struct {
	DB *gorm.DB
}

type FileStat struct {
	Path string
	Size int64
}

func (a *App) initialize(dbURI string) error {
	db, err := gorm.Open(sqlite.Open(dbURI))
	if err != nil {
		return fmt.Errorf("failed to connect database")
	}
	a.DB = db
	return nil
}

// TODO: Set Quota, Set UsedSpace
func (a *App) createClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		fmt.Println("user not authorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Println("failed in ParseForm()")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := &Client{
		Name:     r.PostFormValue("name"),
		AuthCode: generateToken(32),
	}

	a.DB.Create(client)

	storagePath := path.Join(cfg.StoragesPath, client.Name, "chunks", "empty")
	cfgDir := filepath.Dir(storagePath)
	if !utils.Exist(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			fmt.Println("failed to create storage path for client")
			w.WriteHeader(http.StatusInternalServerError)
			a.DB.Delete(client)
			return
		}
	}

	u, err := url.Parse(fmt.Sprintf("/clients/%d", client.ID))
	if err != nil {
		fmt.Println("failed to form a new client URL")
		w.WriteHeader(http.StatusInternalServerError)
		os.RemoveAll(cfgDir)
		a.DB.Delete(client)
		return
	}
	base, err := url.Parse(r.URL.String())
	if err != nil {
		fmt.Println("failed to parse request URL")
		w.WriteHeader(http.StatusInternalServerError)
		os.RemoveAll(cfgDir)
		a.DB.Delete(client)
		return
	}

	w.Header().Set("Location", base.ResolveReference(u).String())
}

func (a *App) getAllClients(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		fmt.Println("user not authorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var clients []Client

	a.DB.Find(&clients)
	clientsJSON, _ := json.Marshal(clients)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(clientsJSON))
}

func (a *App) getClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		fmt.Println("user not authorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var client Client
	vars := mux.Vars(r)

	a.DB.First(&client, "id = ?", vars["id"])
	clientJSON, _ := json.Marshal(client)

	// w.WriteHeader(http.StatusOK)
	w.Write([]byte(clientJSON))
}

func (a *App) getClientByAuthCode(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(w, r)
	if err != nil {
		fmt.Println("client not authorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	fmt.Printf("%+v\n", client)
	clientJSON, _ := json.Marshal(client)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(clientJSON))
}

// TODO: Rename Folder, Remove Authcode changes, Add Quota update
func (a *App) updateClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		fmt.Println("user not authorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	if err := r.ParseForm(); err != nil {
		fmt.Println("failed in ParseForm() call")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := &Client{
		Name:     r.PostFormValue("name"),
		AuthCode: r.PostFormValue("authcode"),
	}

	a.DB.Model(&client).Where("id = ?", vars["id"]).Updates(&client)

	w.WriteHeader(http.StatusNoContent)
}
func (a *App) deleteClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		fmt.Println("user not authorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	var client Client
	a.DB.First(&client, "id = ?", vars["id"])
	os.RemoveAll(path.Join(cfg.StoragesPath, client.Name))

	a.DB.Delete(&Client{}, vars["id"])

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

var (
	serveCmd = &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			a := &App{}
			u, err := utils.PathToUrl(cfgFileName)
			if err != nil {
				return fmt.Errorf("config path isn't a valid url")
			}

			err = cfg.Load(u)
			if err != nil {
				return fmt.Errorf("couldn't load config file")
			}

			err = a.initialize(cfg.DBFileName)
			if err != nil {
				return err
			}

			router := mux.NewRouter()
			// router.Handle("/webui/", http.StripPrefix("/webui/", http.FileServer(http.Dir(filepath.Join("ui", "build")))))
			router.HandleFunc("/login", a.login)
			router.HandleFunc("/clients", a.createClient).Methods("POST")
			router.HandleFunc("/clients", a.getAllClients).Methods("GET", "OPTIONS")
			router.HandleFunc("/clients/{id}", a.getClient).Methods("GET")
			router.HandleFunc("/clients/{id}", a.updateClient).Methods("PUT")
			router.HandleFunc("/clients/{id}", a.deleteClient).Methods("DELETE")
			router.HandleFunc("/testUser", a.testUserAuth).Methods("GET")

			fmt.Println("starting server")
			router.HandleFunc("/upload", a.upload).Methods("POST")
			router.PathPrefix("/download/").HandlerFunc(a.download).Methods("GET")
			// router.HandleFunc("/repository", a.repository)
			// router.HandleFunc("/snapshot", a.uploadSnapshot).Methods("POST")
			// router.HandleFunc("/snapshot/", a.downloadSnapshot).Methods("GET")
			router.PathPrefix("/size/").HandlerFunc(a.getFileStats).Methods("GET")
			router.PathPrefix("/mkdir/").HandlerFunc(a.mkdir).Methods("GET")
			router.PathPrefix("/delete/").HandlerFunc(a.delete).Methods("DELETE")
			router.HandleFunc("/getClientByAuthCode", a.getClientByAuthCode).Methods("GET")
			router.HandleFunc("/testClient", a.testClientAuth).Methods("GET")

			http.Handle("/", router)
			// To show available routes (for development)
			router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
				tpl, err1 := route.GetPathTemplate()
				met, err2 := route.GetMethods()
				fmt.Println(tpl, err1, met, err2)
				return nil
			})
			err = http.ListenAndServe(":"+cfg.AdminUIPort, nil)
			if err != nil {
				return fmt.Errorf("port occupied")
			}
			return nil
		},
	}

	cfgFileName string
)

func init() {
	serveCmd.PersistentFlags().StringVarP(&cfgFileName, "configURL", "C", config.DefaultPath(), "Path to configuration file")
	RootCmd.AddCommand(serveCmd)
}

// curl -H "Authorization: Bearer 9b1610f4cb673feeee90fb9c8cfed2422caa6f6478dee79c3a54b72ffddae1f2" http://localhost:42024/testClient

func (a *App) testClientAuth(w http.ResponseWriter, r *http.Request) {
	if client, err := a.authenticateClient(w, r); err != nil {
		fmt.Fprintf(w, "error")
	} else {
		fmt.Fprintf(w, "Client name is: %s", client.Name)
	}
}

func (a *App) authenticateClient(w http.ResponseWriter, r *http.Request) (*Client, error) {
	authTokenHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")

	if len(authTokenHeader) < 2 {
		return nil, fmt.Errorf("no authorization was given")
	}

	authToken := authTokenHeader[1]

	client := &Client{}
	if err := a.DB.First(client, Client{AuthCode: authToken}).Error; err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return client, nil
}

// Example - username: abc, password: 123
// Encode username and password, format for BasicAuth: "username:password", -n flag for echo needs to be set, to get rid of the \n
// echo -n "abc:123" | base64
// output: YWJjOjEyMw==
// Set Header and you are good to go
// curl -H "Authorization: Basic YWJjOjEyMw==" http://localhost:42024/testUser

func (a *App) testUserAuth(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		fmt.Fprintf(w, "error")
	} else {
		fmt.Fprintf(w, "User name authenticated")
	}
}

func (a *App) authenticateUser(w http.ResponseWriter, r *http.Request) error {
	u, p, ok := r.BasicAuth()

	if !ok {
		return fmt.Errorf("security alert: no auth set")
	}

	if u != cfg.AdminUserName || utils.CheckPasswordHash(p, cfg.AdminPassword) {
		return fmt.Errorf("security alert: authentication failed")
	}

	return nil
}

// upload logic.
func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving upload")

	client, err := a.authenticateClient(w, r)

	if r.Method != "POST" || err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlPath := r.Header.Get("Path")
	if len(r.URL.Path) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fileContent := r.FormValue("uploadfile")
	if err = upload(*a, *client, urlPath, fileContent); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func upload(a App, client Client, filePath string, fileContent string) error {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", filePath))
	if err != nil || strings.Contains(filePath, "..") {
		return fmt.Errorf("invalid url")
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if (client.Quota - client.UsedSpace) < uint64(len([]byte(fileContent))) {
		return fmt.Errorf("client storage space used up")
	}

	_, err = io.Copy(f, bytes.NewReader([]byte(fileContent)))
	if err != nil {
		return err
	}
	stats, err := os.Stat(path)
	if err != nil {
		defer os.Remove(path)
		return err
	}
	client.UsedSpace += uint64(stats.Size())
	a.DB.Model(&client).Where("id = ?", client.ID).Updates(&client)
	return nil
}

// download logic.
func (a *App) download(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(w, r)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if len(r.URL.Path) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	path, err := downloadFile(*client, r.URL.Path[10:])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		http.ServeFile(w, r, filepath.Join(path))
		return
	}

	w.WriteHeader(http.StatusBadRequest)
}

func downloadFile(client Client, filePath string) (string, error) {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", filePath))
	if err != nil || strings.Contains(filePath, "..") {
		return "", fmt.Errorf("invalid url")
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))
	if utils.Exist(path) {
		return path, nil
	}
	return "", fmt.Errorf("path not found")
}

func (a *App) getFileStats(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Getting status of file")

	client, err := a.authenticateClient(w, r)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if len(r.URL.Path) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, err := stat(*client, r.URL.Path[6:])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jData, err := json.Marshal(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jData))
}

func stat(client Client, filePath string) (FileStat, error) {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", filePath))
	if err != nil || strings.Contains(filePath, "..") {
		return FileStat{}, fmt.Errorf("ERROR: Invalid url")
	}
	var file FileStat
	file.Path = filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))

	osFile, err := os.Stat(file.Path)
	if err != nil {
		return FileStat{}, fmt.Errorf("ERROR: %v", err)
	}
	file.Size = osFile.Size()

	return file, nil
}

func (a *App) mkdir(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(w, r)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	fmt.Printf("%v\n", r.URL.Path)
	if len(r.URL.Path) < 7 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = mkDir(*client, r.URL.Path[7:])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func mkDir(client Client, dirPath string) error {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", dirPath))
	if err != nil || strings.Contains(dirPath, "..") {
		return fmt.Errorf("ERROR: Invalid url")
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", dirPath))

	if !utils.Exist(path) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("ERROR: %v", err)
		}
	}

	return nil
}

func (a *App) delete(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(w, r)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if len(r.URL.Path) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = deleteFile(*a, *client, r.URL.Path[8:])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteFile(a App, client Client, filePath string) error {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", filePath))
	if err != nil || strings.Contains(filePath, "..") {
		return fmt.Errorf("ERROR: Invalid url")
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))

	stats, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("ERROR: %v", err)
	}

	if stats.IsDir() {
		return fmt.Errorf("ERROR: can't delete folders")
	}

	client.UsedSpace -= uint64(stats.Size())
	a.DB.Model(&client).Where("id = ?", client.ID).Updates(&client)

	if utils.Exist(path) {
		if err := os.Remove(path); err != nil {
			client.UsedSpace += uint64(stats.Size())
			return fmt.Errorf("ERROR: %v", err)
		}
	}

	return nil
}

func AvailableSpace() (uint64, error) {
	statOS := &StatOS{}
	space, err := statOS.GetAvailableStorageSpace()
	if err != nil {
		return 0, err
	}

	return space, nil
}
