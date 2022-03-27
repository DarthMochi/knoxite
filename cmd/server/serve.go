//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

	quota, err := strconv.ParseUint(r.PostFormValue("quota"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	availableSpace, err := a.AvailableSpace()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if quota > availableSpace {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name := r.PostFormValue("name")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := &Client{
		Name:     name,
		Quota:    quota,
		AuthCode: generateToken(32),
	}

	if strings.Contains(client.Name, "..") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.DB.Create(client)

	storagePath := filepath.Join(cfg.StoragesPath, client.Name, "chunks", "empty")
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

	w.WriteHeader(http.StatusOK)
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

	var oldClient Client
	a.DB.First(&oldClient, vars["id"])

	oldName := oldClient.Name

	quota, err := strconv.ParseUint(r.PostFormValue("quota"), 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	availableSpace, err := a.AvailableSpace()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if quota > availableSpace || quota < oldClient.UsedSpace {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	client := &Client{
		Name:  r.PostFormValue("name"),
		Quota: quota,
	}

	if strings.Contains(client.Name, "..") {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = os.Rename(filepath.Join("/", cfg.StoragesPath, oldName), filepath.Join("/", cfg.StoragesPath, client.Name))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.DB.Model(&client).Where("id = ?", vars["id"]).Updates(&client)

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) deleteClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	var client Client
	a.DB.First(&client, "id = ?", vars["id"])
	os.RemoveAll(filepath.Join(cfg.StoragesPath, client.Name))

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
			router.PathPrefix("/size/").HandlerFunc(a.getFileStats).Methods("GET")
			router.PathPrefix("/mkdir/").HandlerFunc(a.mkdir).Methods("GET")
			router.PathPrefix("/delete").HandlerFunc(a.delete).Methods("DELETE")
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
	if _, err := a.authenticateClient(w, r); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
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
		w.WriteHeader(http.StatusUnauthorized)
	} else {
		w.WriteHeader(http.StatusOK)
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
	diff, err := a.UploadFile(*client, urlPath, fileContent)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	filestat := FileStat{
		Path: urlPath,
		Size: diff,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(filestat)
}

func (a *App) UploadFile(client Client, filePath string, fileContent string) (int64, error) {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", filePath))
	if err != nil || strings.Contains(filePath, "..") {
		return 0, fmt.Errorf("invalid url")
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	if (client.Quota - client.UsedSpace) < uint64(len([]byte(fileContent))) {
		return 0, fmt.Errorf("client storage space used up")
	}

	fileinfo, err := f.Stat()
	if err != nil {
		return 0, fmt.Errorf("couldn't read file information")
	}

	buf := make([]byte, fileinfo.Size())
	f.Read(buf)
	diff, err := utils.ByteArrDiff(buf, []byte(fileContent))
	if err != nil {
		return 0, err
	}

	err = os.WriteFile(path, []byte(fileContent), 0600)
	if err != nil {
		return 0, err
	}
	if utils.Abs(int64(client.UsedSpace)-diff) < 0 {
		client.UsedSpace = 0
	} else if diff < 0 {
		client.UsedSpace -= uint64(diff)
	} else if diff > 0 {
		client.UsedSpace += uint64(diff)
	}
	a.DB.Model(&client).Where("id = ?", client.ID).Updates(&client)
	return diff, nil
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

	fmt.Println("Serving file: ", r.URL.Path[10:])
	data, err := a.DownloadFile(*client, r.URL.Path[10:])
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (a *App) DownloadFile(client Client, filePath string) ([]byte, error) {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", filePath))
	if err != nil || strings.Contains(filePath, "..") {
		return nil, fmt.Errorf("invalid url")
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))
	if !utils.Exist(path) {
		return nil, fmt.Errorf("path not found")
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0600)

	return ioutil.ReadAll(f)
}

func (a *App) getFileStats(w http.ResponseWriter, r *http.Request) {
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

	urlPath := r.Header.Get("Path")
	if len(urlPath) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = a.DeleteFile(*client, urlPath)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *App) DeleteFile(client Client, filePath string) error {
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

	if utils.Exist(path) {
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("ERROR: %v", err)
		}
	}

	client.UsedSpace -= uint64(stats.Size())
	a.DB.Model(&client).Where("id = ?", client.ID).Updates(&client)

	return nil
}

func (a *App) AvailableSpace() (uint64, error) {
	statOS := &StatOS{}
	space, err := statOS.GetAvailableStorageSpace()
	if err != nil {
		return 0, err
	}

	var totalQuota uint64
	a.DB.Table("clients").Select("sum(quota)").Row().Scan(&totalQuota)

	return space - totalQuota, nil
}
