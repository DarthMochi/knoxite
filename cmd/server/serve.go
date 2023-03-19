//go:build windows || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build windows darwin dragonfly freebsd linux netbsd openbsd solaris

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/knoxite/knoxite/cmd/server/config"
	"github.com/knoxite/knoxite/cmd/server/utils"
	"github.com/miekg/dns"
	"github.com/rs/cors"
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

type ServerConfig struct {
	Port         string
	Hostname     string
	ServerScheme string
}

func makeHTTPServer(handler http.Handler, port string) *http.Server {
	return &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      handler,
		Addr:         ":" + port,
	}
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
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	base, err := url.Parse(r.URL.String())
	if err != nil {
		WarningLogger.Println(errInvalidURL)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.ParseForm(); err != nil {
		WarningLogger.Println(errInvalidBody)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	quota := r.PostFormValue("quota")
	name := r.PostFormValue("name")

	u, err := a.NewClient(name, quota)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Location", base.ResolveReference(u).String())
}

func (a *App) getAllClients(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
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
	if _, err := w.Write(clientsJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) getClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var client Client
	vars := mux.Vars(r)

	a.DB.First(&client, "id = ?", vars["id"])
	clientJSON, _ := json.Marshal(client)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(clientJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) getClientByName(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var client Client
	vars := mux.Vars(r)

	a.DB.First(&client, "name = ?", vars["name"])
	clientJSON, _ := json.Marshal(client)

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(clientJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) updateClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
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

	clientId := vars["id"]
	name := r.PostFormValue("name")
	quota := r.PostFormValue("quota")

	if err := a.UpdateClient(clientId, name, quota); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) deleteClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
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
	if err := a.authenticateUser(r); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (a *App) totalAvailableSpace(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
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
	if _, err := w.Write(sizeJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) totalAvailableSpaceMinusQuota(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	size, err := a.AvailableSpaceMinusQuota()
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
	if _, err := w.Write(sizeJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) totalAvailableSpacePlusQuota(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)

	var client Client
	a.DB.First(&client, "id = ?", vars["id"])

	size, err := a.AvailableSpacePlusQuota(client.Quota)
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
	if _, err := w.Write(sizeJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) totalUsedSpace(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	size, err := a.UsedSpace()
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
	if _, err := w.Write(sizeJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) totalOccupiedQuota(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	size, err := a.TotalQuota()
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
	if _, err := w.Write(sizeJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) configurationInformation(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(r); err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var config ServerConfig
	config.Port = cfg.HTTPPort
	config.Hostname = cfg.AdminHostname
	config.ServerScheme = "http"

	if cfg.UseHTTPS {
		config.ServerScheme = "https"
		config.Port = cfg.HTTPSPort
	}

	configJSON, err := json.Marshal(config)
	if err != nil {
		WarningLogger.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(configJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	path := filepath.Clean(r.URL.Path)
	if path == "/" || ((strings.HasPrefix(path, "/admin") || strings.HasPrefix(path, "/login")) && filepath.Ext(path) == "") {
		path = "index.html"
	}
	path = strings.TrimPrefix(path, "/")

	file, err := uiFS.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			InfoLogger.Println("file", path, "not found:", err)
			http.NotFound(w, r)
			return
		}
		InfoLogger.Println("file", path, "connot be read:", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(path))
	w.Header().Set("Content-Type", contentType)
	if strings.HasPrefix(path, "static/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	}

	stat, err := file.Stat()
	if err == nil && stat.Size() > 0 {
		w.Header().Set("Content-length", fmt.Sprintf("%d", stat.Size()))
	}

	n, _ := io.Copy(w, file)
	InfoLogger.Println("file", path, "copied", n, "bytes")
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

			router.HandleFunc("/api/login", a.login).Methods("POST", "OPTIONS")
			router.HandleFunc("/api/clients", a.createClient).Methods("POST")
			router.HandleFunc("/api/clients", a.getAllClients).Methods("GET", "OPTIONS")
			router.HandleFunc("/api/clients/{id}", a.getClient).Methods("GET")
			router.HandleFunc("/api/clients/{name}", a.getClientByName).Methods("GET")
			router.HandleFunc("/api/clients/{id}", a.updateClient).Methods("PUT")
			router.HandleFunc("/api/clients/{id}", a.deleteClient).Methods("DELETE")
			router.HandleFunc("/api/storage_size", a.totalAvailableSpace).Methods("GET")
			router.HandleFunc("/api/storage_size_minus_quota", a.totalAvailableSpaceMinusQuota).Methods("GET")
			router.HandleFunc("/api/storage_size_plus_quota", a.totalAvailableSpacePlusQuota).Methods("GET")
			router.HandleFunc("/api/used_space", a.totalUsedSpace).Methods("GET")
			router.HandleFunc("/api/total_quota", a.totalOccupiedQuota).Methods("GET")
			router.HandleFunc("/api/server_config", a.configurationInformation).Methods("GET")

			router.HandleFunc("/upload", a.upload).Methods("POST")
			router.PathPrefix("/download/").HandlerFunc(a.download).Methods("GET")
			router.PathPrefix("/stat/").HandlerFunc(a.getFileStats).Methods("GET")
			router.PathPrefix("/mkdir/").HandlerFunc(a.mkdir).Methods("GET")
			router.PathPrefix("/delete").HandlerFunc(a.delete).Methods("DELETE")
			router.HandleFunc("/getClientByAuthCode", a.getClientByAuthCode).Methods("GET")
			router.HandleFunc("/download_cert", a.downloadCert).Methods("GET")

			// router.HandleFunc("/", a.handleStatic).Methods("GET")
			// fsWebUI := http.FileServer(uiFS.Open())
			router.PathPrefix("/").HandlerFunc(a.handleStatic).Methods("GET")
			// if os.Getenv("APP_ENV") == "production" {
			// 	fsWebUI := http.FileServer(http.Dir(filepath.Join(uiPath, "build")))
			// 	router.PathPrefix("/").Handler(http.StripPrefix("/", fsWebUI)).Methods("GET")
			// }
			router.Use(loggingMiddleware)
			http.Handle("/", router)
			err = router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
				tpl, err1 := route.GetPathTemplate()
				met, err2 := route.GetMethods()
				InfoLogger.Println(tpl, err1, met, err2)
				return nil
			})
			if err != nil {
				InfoLogger.Println(err)
			}

			if cfg.UseHTTPS {
				certPem := filepath.Join(certsPath, "knoxite-server-cert.pem")
				keyPem := filepath.Join(certsPath, "knoxite-server-key.pem")
				go func() {
					c := cors.New(cors.Options{
						AllowedOrigins:   []string{"*"},
						AllowCredentials: true,
					})
					httpsSrv := makeHTTPServer(c.Handler(router), cfg.HTTPSPort)

					err = httpsSrv.ListenAndServeTLS(certPem, keyPem)
					if err != nil {
						ErrorLogger.Println("httpsSrc.ListenAndServeTLS() failed with " + err.Error())
					}
				}()

			}

			c := cors.New(cors.Options{
				AllowedOrigins:   []string{"*"},
				AllowCredentials: true,
			})
			httpSrv := makeHTTPServer(c.Handler(router), cfg.HTTPPort)
			err = httpSrv.ListenAndServe()
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

var uiFS fs.FS

func init() {
	var err error
	if err = setPaths(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}

	uiFS, err = fs.Sub(UI, "ui/build")
	if err != nil {
		fmt.Printf("failed to get ui fs: %v\n", err)
		os.Exit(-1)
	}
	serveCmd.PersistentFlags().StringVarP(&cfgFileName, "configURL", "C", config.DefaultPath(), "Path to configuration file")
	RootCmd.AddCommand(serveCmd)
}

func (a *App) getClientByAuthCode(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(r)
	if err != nil {
		WarningLogger.Println(errNoAuth)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	clientJSON, _ := json.Marshal(client)
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(clientJSON); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) authenticateClient(r *http.Request) (*Client, error) {
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

func (a *App) authenticateUser(r *http.Request) error {
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
	client, err := a.authenticateClient(r)

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
	if err := json.NewEncoder(w).Encode(filestat); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (a *App) UploadFile(client Client, filePath string, fileContent string) (int64, error) {
	_, err := filepath.Rel(filepath.Join("/", cfg.StoragesPath, client.Name), filepath.Join("/", filePath))
	if err != nil || strings.Contains(filePath, "..") {
		WarningLogger.Println(errInvalidURL)
		return 0, errInvalidURL
	}
	path := filepath.Join("/", cfg.StoragesPath, client.Name, filepath.Join("/", filePath))
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		WarningLogger.Println(errNoSpace)
		return 0, err
	}

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
	_, err = f.Read(buf)
	if err != nil {
		WarningLogger.Println(err)
		return 0, err
	}
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

func (a *App) downloadCert(w http.ResponseWriter, r *http.Request) {
	_, err := a.authenticateClient(r)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	data, err := os.ReadFile(filepath.Join(certsPath, "knoxite-server-cert.pem"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// download logic.
func (a *App) download(w http.ResponseWriter, r *http.Request) {
	client, err := a.authenticateClient(r)

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
	if _, err := w.Write(data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
	client, err := a.authenticateClient(r)
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
	if _, err := w.Write(jData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
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
	client, err := a.authenticateClient(r)
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
	client, err := a.authenticateClient(r)
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

func (a *App) AvailableSpacePlusQuota(quota uint64) (uint64, error) {
	space, err := a.AvailableSpaceMinusQuota()
	if err != nil {
		return 0, err
	}

	return space + quota, nil
}

func (a *App) AvailableSpaceMinusQuota() (uint64, error) {
	space, err := a.AvailableSpace()
	if err != nil {
		return 0, err
	}

	total_quota, err := a.TotalQuota()
	if err != nil {
		return 0, err
	}

	return space - total_quota, nil
}

func (a *App) AvailableSpace() (uint64, error) {
	statOS := &StatOS{}
	space, err := statOS.GetAvailableStorageSpace()
	if err != nil {
		return 0, err
	}

	return space, nil
}

func (a *App) TotalQuota() (uint64, error) {
	var totalQuota sql.NullString
	if err := a.DB.Table("clients").Select("sum(quota)").Row().Scan(&totalQuota); err != nil {
		return 0, err
	}

	var quota uint64
	quota, err := strconv.ParseUint(totalQuota.String, 10, 64)
	if err != nil {
		return 0, nil
	}

	return quota, nil
}

func (a *App) UsedSpace() (uint64, error) {
	var usedSpace uint64
	if err := a.DB.Table("clients").Select("sum(used_space)").Row().Scan(&usedSpace); err != nil {
		return 0, err
	}

	return usedSpace, nil
}
