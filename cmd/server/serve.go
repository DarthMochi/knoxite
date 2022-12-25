//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/knoxite/knoxite/cmd/server/config"
	"github.com/knoxite/knoxite/cmd/server/utils"
	"github.com/miekg/dns"
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
		return err
	}
	a.DB = db
	return nil
}

func (a *App) createClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		WarningLogger.Println(errInvalidBody)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	quota, err := strconv.ParseUint(r.PostFormValue("quota"), 10, 64)
	if err != nil {
		WarningLogger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	availableSpace, err := a.AvailableSpace()
	if err != nil {
		WarningLogger.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if quota > availableSpace {
		WarningLogger.Println(errNoSpace)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	name := r.PostFormValue("name")

	client := &Client{
		Name:     name,
		Quota:    quota,
		AuthCode: generateToken(32),
	}

	if strings.Contains(client.Name, "..") {
		WarningLogger.Println(errInvalidURL)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	a.DB.Create(client)

	storagePath := filepath.Join(cfg.StoragesPath, client.Name)
	cfgDir := filepath.Dir(storagePath)
	if !utils.Exist(cfgDir) {
		if err := os.MkdirAll(cfgDir, 0755); err != nil {
			WarningLogger.Println(errNoPath)
			w.WriteHeader(http.StatusInternalServerError)
			a.DB.Delete(client)
			return
		}
	}

	u, err := utils.ParseClientURL(client.ID)
	if err != nil {
		WarningLogger.Println(errInvalidURL)
		w.WriteHeader(http.StatusInternalServerError)
		os.RemoveAll(cfgDir)
		a.DB.Delete(client)
		return
	}
	base, err := url.Parse(r.URL.String())
	if err != nil {
		WarningLogger.Println(errInvalidURL)
		w.WriteHeader(http.StatusInternalServerError)
		os.RemoveAll(cfgDir)
		a.DB.Delete(client)
		return
	}

	w.Header().Set("Location", base.ResolveReference(u).String())
}

func (a *App) getAllClients(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		WarningLogger.Println(errNoAuth)
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
		WarningLogger.Println(errNoAuth)
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

func (a *App) getClientByName(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var client Client
	vars := mux.Vars(r)

	a.DB.First(&client, "name = ?", vars["name"])
	clientJSON, _ := json.Marshal(client)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(clientJSON))
}

func (a *App) getClientByAuthCode(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(w, r)
	if err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	clientJSON, _ := json.Marshal(client)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(clientJSON))
}

func (a *App) updateClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	if err := r.ParseForm(); err != nil {
		WarningLogger.Println(errInvalidBody)
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

	cfgDir := filepath.Dir(filepath.Join("/", cfg.StoragesPath, client.Name))
	if !utils.Exist(cfgDir) {
		err = os.Rename(filepath.Join("/", cfg.StoragesPath, oldName), filepath.Join("/", cfg.StoragesPath, client.Name))
		if err != nil {
			WarningLogger.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
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

	a.DB.Unscoped().Delete(&client)

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *App) totalAvailableStorageSize(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	size, err := a.AvailableSpace()
	if err != nil {
		WarningLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sizeJSON, err := json.Marshal(size)
	if err != nil {
		WarningLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sizeJSON))
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
			InitLogger()
			a := &App{}
			u, err := utils.PathToUrl(cfgFileName)
			if err != nil {
				ErrorLogger.Println("config path isn't a valid url")
				return err
			}

			err = cfg.Load(u)
			if err != nil {
				ErrorLogger.Println("couldn't load config file")
				return err
			}

			err = a.initialize(cfg.DBFileName)
			if err != nil {
				ErrorLogger.Println("failed to connect database")
				return err
			}

			if cfg.UseHostname {
				hostname, err := os.Hostname()
				if err != nil {
					ErrorLogger.Println("hostname couldn't be retrieved")
					return err
				}

				r := new(dns.A)
				r.Hdr = dns.RR_Header{Name: hostname + ".", Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 3600}
				r.A = utils.GetLocalIP()
				if err != nil {
					ErrorLogger.Println("dns couldn't be setup")
					return err
				}
			}

			InfoLogger.Println("starting server")
			router := mux.NewRouter()

			router.HandleFunc("/login", a.login).Methods("POST")
			router.HandleFunc("/clients", a.createClient).Methods("POST")
			router.HandleFunc("/clients", a.getAllClients).Methods("GET", "OPTIONS")
			router.HandleFunc("/clients/{id}", a.getClient).Methods("GET")
			router.HandleFunc("/clients/{name}", a.getClientByName).Methods("GET")
			router.HandleFunc("/clients/{id}", a.updateClient).Methods("PUT")
			router.HandleFunc("/clients/{id}", a.deleteClient).Methods("DELETE")
			router.HandleFunc("/storage_size", a.totalAvailableStorageSize).Methods("GET")

			router.HandleFunc("/upload", a.upload).Methods("POST")
			router.PathPrefix("/download/").HandlerFunc(a.download).Methods("GET")
			router.PathPrefix("/stat/").HandlerFunc(a.getFileStats).Methods("GET")
			router.PathPrefix("/mkdir/").HandlerFunc(a.mkdir).Methods("GET")
			router.PathPrefix("/delete").HandlerFunc(a.delete).Methods("DELETE")
			router.HandleFunc("/getClientByAuthCode", a.getClientByAuthCode).Methods("GET")
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			if os.Getenv("APP_ENV") == "production" {
				fsWebUI := http.FileServer(http.Dir(filepath.Join(wd, "cmd", "server", "ui", "build")))
				router.PathPrefix("/").Handler(http.StripPrefix("/", fsWebUI)).Methods("GET")
			}
			router.Use(loggingMiddleware)
			http.Handle("/", router)
			router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
				tpl, err1 := route.GetPathTemplate()
				met, err2 := route.GetMethods()
				InfoLogger.Println(tpl, err1, met, err2)
				return nil
			})
			if cfg.UseHTTPS {
				certsDir := filepath.Join(wd, "cmd", "server", "certs")
				certPem := filepath.Join(certsDir, "cert.pem")
				keyPem := filepath.Join(certsDir, "key.pem")
				err = http.ListenAndServeTLS(":"+cfg.AdminUIPort, certPem, keyPem, router)
			} else {
				err = http.ListenAndServe(":"+cfg.AdminUIPort, nil)
			}
			if err != nil {
				ErrorLogger.Println("port occupied")
				return err
			}
			return nil
		},
	}

	cfgFileName string
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		InfoLogger.Println(utils.LogRequest(r))
		next.ServeHTTP(w, r)
	})
}

func init() {
	serveCmd.PersistentFlags().StringVarP(&cfgFileName, "configURL", "C", config.DefaultPath(), "Path to configuration file")
	RootCmd.AddCommand(serveCmd)
}

func (a *App) authenticateClient(w http.ResponseWriter, r *http.Request) (*Client, error) {
	authTokenHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")

	if len(authTokenHeader) < 2 {
		WarningLogger.Println(errNoAuth)
		return nil, errNoAuth
	}

	authToken := authTokenHeader[1]

	client := &Client{}
	if err := a.DB.First(client, Client{AuthCode: authToken}).Error; err != nil {
		WarningLogger.Println(err)
		return nil, err
	}
	return client, nil
}

func (a *App) authenticateUser(w http.ResponseWriter, r *http.Request) error {
	u, p, ok := r.BasicAuth()

	if !ok {
		WarningLogger.Println(errNoAuth)
		return errNoAuth
	}

	if u != cfg.AdminUserName || utils.CheckPasswordHash(p, cfg.AdminPassword) {
		WarningLogger.Println(errNoAuth)
		return errNoAuth
	}

	return nil
}

// upload logic.
func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(w, r)

	if r.Method != "POST" || err != nil {
		WarningLogger.Println(errInvalidURL)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlPath := r.Header.Get("Path")
	if len(r.URL.Path) < 1 {
		WarningLogger.Println(errInvalidHeader)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		WarningLogger.Println(errInvalidBody)
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
		WarningLogger.Println(errInvalidURL)
		return 0, errInvalidURL
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		WarningLogger.Println(err)
		return 0, err
	}
	defer f.Close()

	if (client.Quota - client.UsedSpace) < uint64(len([]byte(fileContent))) {
		WarningLogger.Println(errNoSpace)
		return 0, errNoSpace
	}

	fileinfo, err := f.Stat()
	if err != nil {
		WarningLogger.Println(errInvalidBody)
		return 0, err
	}

	buf := make([]byte, fileinfo.Size())
	f.Read(buf)
	diff := utils.ByteArrDiff(buf, []byte(fileContent))
	if diff == 0 {
		return diff, nil
	}

	err = os.WriteFile(path, []byte(fileContent), 0600)
	if err != nil {
		WarningLogger.Println(err)
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
		WarningLogger.Println(errInvalidURL)
		return nil, errInvalidURL
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))
	if !utils.Exist(path) {
		WarningLogger.Println(errNoPath)
		return nil, errNoPath
	}

	f, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		WarningLogger.Println(err)
		return nil, err
	}

	return io.ReadAll(f)
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
		WarningLogger.Println(errInvalidURL)
		return FileStat{}, errInvalidURL
	}
	var file FileStat
	file.Path = filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))

	osFile, err := os.Stat(file.Path)
	if err != nil {
		WarningLogger.Println(err)
		return FileStat{}, err
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
		WarningLogger.Println(errInvalidURL)
		return errInvalidURL
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", dirPath))

	if !utils.Exist(path) {
		if err := os.MkdirAll(path, 0755); err != nil {
			WarningLogger.Println(err)
			return err
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
		WarningLogger.Println(errInvalidURL)
		return errInvalidURL
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))

	stats, err := os.Stat(path)
	if err != nil {
		WarningLogger.Println(err)
		return err
	}

	if stats.IsDir() {
		WarningLogger.Println(errNoPath)
		return errNoPath
	}

	if utils.Exist(path) {
		if err := os.Remove(path); err != nil {
			WarningLogger.Println(errCantDelete)
			return errCantDelete
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
