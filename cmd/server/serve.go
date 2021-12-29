package main

import (
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
	client := &Client{
		Name: r.PostFormValue("name"),
		// AuthCode: r.PostFormValue("authcode"),
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
	w.WriteHeader(http.StatusCreated)
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

// TODO: Rename Folder, Remove Authcode changes....
func (a *App) updateClient(w http.ResponseWriter, r *http.Request) {
	if err := a.authenticateUser(w, r); err != nil {
		fmt.Println("user not authorized")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)

	if err := r.ParseForm(); err != nil {
		panic("failed in ParseForm() call")
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
			router.HandleFunc("/repository", a.repository)
			router.HandleFunc("/snapshot", a.uploadSnapshot).Methods("POST")
			router.HandleFunc("/snapshot/", a.downloadSnapshot).Methods("GET")
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
		fmt.Errorf("error")
	} else {
		fmt.Fprintf(w, "Client name is: %s", client.Name)
	}
}

func (a *App) authenticateClient(w http.ResponseWriter, r *http.Request) (*Client, error) {
	authTokenHeader := strings.Split(r.Header.Get("Authorization"), "Bearer ")

	if len(authTokenHeader) < 2 {
		return nil, fmt.Errorf("No authorization was given")
	}

	authToken := authTokenHeader[1]

	client := &Client{}
	if err := a.DB.First(client, Client{AuthCode: authToken}).Error; err != nil {
		// respondError(w, http.StatusNotFound, err.Error())
		w.WriteHeader(http.StatusNotFound)
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
		fmt.Errorf("error")
	} else {
		fmt.Fprintf(w, "User name authenticated")
	}
}

func (a *App) authenticateUser(w http.ResponseWriter, r *http.Request) error {
	u, p, ok := r.BasicAuth()
	fmt.Printf("Username: %s\n", u)
	fmt.Printf("Password: %s\n", p)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return fmt.Errorf("security alert: no auth set")
	}

	if u != cfg.AdminUserName || utils.CheckPasswordHash(p, cfg.AdminPassword) {
		w.WriteHeader(http.StatusUnauthorized)
		return fmt.Errorf("security alert: authentication failed")
	}

	w.WriteHeader(http.StatusOK)
	return nil
}

// upload logic.
func (a *App) upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving upload")

	client, err := a.authenticateClient(w, r)

	if r.Method == "POST" && err == nil {
		fmt.Printf("Client: %s\n", client.Name)
		path := filepath.Join(cfg.StoragesPath, client.Name)

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer file.Close()

		fmt.Fprintf(w, "%v", handler.Header)
		f, err := os.OpenFile(filepath.Join(path, "chunks", handler.Filename), os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer f.Close()

		_, err = io.Copy(f, file)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Println("Stored chunk", filepath.Join(path, "chunks", handler.Filename))
	}
}

// download logic.
func (a *App) download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving chunk", r.URL.Path[10:])

	/* path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}*/
	client, err := a.authenticateClient(w, r)

	if err != nil {
		return
	}
	path := filepath.Join(cfg.StoragesPath, client.Name)

	if r.Method == "GET" {
		http.ServeFile(w, r, filepath.Join(path, "chunks", r.URL.Path[10:]))
	}
}

// uploadRepo logic.
func (a *App) uploadRepo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving repository")

	client, err := a.authenticateClient(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile(filepath.Join(cfg.StoragesPath, client.Name, "repository.knoxite"), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("Stored repository", filepath.Join(cfg.StoragesPath, client.Name, "repository.knoxite"))
}

// downloadRepo logic.
func (a *App) downloadRepo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving repository")

	client, err := a.authenticateClient(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	http.ServeFile(w, r, filepath.Join(cfg.StoragesPath, client.Name, "repository.knoxite"))
}

func (a *App) repository(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		a.downloadRepo(w, r)
	case "POST":
		a.uploadRepo(w, r)
	}
}

// uploadSnapshot logic.
func (a *App) uploadSnapshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving snapshot")

	client, err := a.authenticateClient(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("uploadfile")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	fmt.Fprintf(w, "%v", handler.Header)
	f, err := os.OpenFile(filepath.Join(cfg.StoragesPath, client.Name, "snapshots", handler.Filename), os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Println("Stored snapshot", filepath.Join(cfg.StoragesPath, client.Name, "snapshots", handler.Filename))
}

// downloadRepo logic.
func (a *App) downloadSnapshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving snapshot", r.URL.Path[10:])

	client, err := a.authenticateClient(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	http.ServeFile(w, r, filepath.Join(cfg.StoragesPath, client.Name, "snapshots", r.URL.Path[10:]))
}
