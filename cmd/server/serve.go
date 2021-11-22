package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
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
	if err := r.ParseForm(); err != nil {
		panic("failed in ParseForm()")
	}
	client := &Client{
		Name:     r.PostFormValue("name"),
		AuthCode: r.PostFormValue("authcode"),
	}
	a.DB.Create(client)

	u, err := url.Parse(fmt.Sprintf("/clients/%d", client.ID))
	if err != nil {
		panic("failed to form a new client URL")
	}
	base, err := url.Parse(r.URL.String())
	if err != nil {
		panic("failed to parse request URL")
	}
	w.Header().Set("Location", base.ResolveReference(u).String())
	w.WriteHeader(201)
}

func (a *App) getAllClients(w http.ResponseWriter, r *http.Request) {
	var clients []Client

	a.DB.Find(&clients)
	clientsJSON, _ := json.Marshal(clients)

	w.WriteHeader(200)
	w.Write([]byte(clientsJSON))
}

func (a *App) getClient(w http.ResponseWriter, r *http.Request) {
	var client Client
	vars := mux.Vars(r)

	a.DB.First(&client, "id = ?", vars["id"])
	clientJSON, _ := json.Marshal(client)

	w.WriteHeader(200)
	w.Write([]byte(clientJSON))
}

func (a *App) updateClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	if err := r.ParseForm(); err != nil {
		panic("failed in ParseForm() call")
	}

	client := &Client{
		Name:     r.PostFormValue("name"),
		AuthCode: r.PostFormValue("authcode"),
	}

	a.DB.Model(&client).Where("id = ?", vars["id"]).Updates(&client)

	w.WriteHeader(204)
}
func (a *App) deleteClient(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	a.DB.Delete(&Client{}, vars["id"])

	w.WriteHeader(204)
}

var serveCmd = &cobra.Command{
	Use: "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := &App{}
		err := a.initialize("test.db")
		if err != nil {
			return err
		}

		router := mux.NewRouter()
		router.HandleFunc("/clients", a.createClient).Methods("POST")
		router.HandleFunc("/clients", a.getAllClients).Methods("GET")
		router.HandleFunc("/clients/{id}", a.getClient).Methods("GET")
		router.HandleFunc("/clients/{id}", a.updateClient).Methods("PUT")
		router.HandleFunc("/clients/{id}", a.deleteClient).Methods("DELETE")

		fmt.Println("starting server")
		/* http.HandleFunc("/test", test)
		http.HandleFunc("/upload", upload)
		http.HandleFunc("/download/", download)
		http.HandleFunc("/repository", repository)
		http.HandleFunc("/snapshot", uploadSnapshot)
		http.HandleFunc("/snapshot/", downloadSnapshot) */

		http.Handle("/", router)
		err = http.ListenAndServe(":42024", nil)
		if err != nil {
			return fmt.Errorf("port occupied")
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
}

const storagePath = "/tmp/knoxite.storage"

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "URL is: %v", r.URL.Path)
}

func authPath(w http.ResponseWriter, r *http.Request) (string, error) {
	auth, _, ok := r.BasicAuth()
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Security alert: no auth set")
	}

	// check for relative path attacks
	if strings.Contains(r.URL.Path, ".."+string(os.PathSeparator)) {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Security alert: url path tampering")
	}
	if strings.Contains(auth, ".."+string(os.PathSeparator)) {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Security alert: auth code tampering")
	}

	dir := filepath.Join(storagePath, auth)
	src, err := os.Stat(dir)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Invalid auth code: unknown user")
	}
	if !src.IsDir() {
		w.WriteHeader(http.StatusUnauthorized)
		return "", errors.New("Invalid auth code: not a dir")
	}

	return dir, nil
}

// upload logic.
func upload(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving upload")
	if r.Method == "POST" {
		path := filepath.Join(storagePath)

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
func download(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving chunk", r.URL.Path[10:])

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	if r.Method == "GET" {
		http.ServeFile(w, r, filepath.Join(path, "chunks", r.URL.Path[10:]))
	}
}

// uploadRepo logic.
func uploadRepo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving repository")

	path, err := authPath(w, r)
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
	f, err := os.OpenFile(filepath.Join(path, "repository.knoxite"), os.O_WRONLY|os.O_CREATE, 0600)
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

	fmt.Println("Stored repository", filepath.Join(path, "repository.knoxite"))
}

// downloadRepo logic.
func downloadRepo(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving repository")

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	http.ServeFile(w, r, filepath.Join(path, "repository.knoxite"))
}

func repository(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		downloadRepo(w, r)
	case "POST":
		uploadRepo(w, r)
	}
}

// uploadSnapshot logic.
func uploadSnapshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Receiving snapshot")

	path, err := authPath(w, r)
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
	f, err := os.OpenFile(filepath.Join(path, "snapshots", handler.Filename), os.O_WRONLY|os.O_CREATE, 0600)
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

	fmt.Println("Stored snapshot", filepath.Join(path, "snapshots", handler.Filename))
}

// downloadRepo logic.
func downloadSnapshot(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Serving snapshot", r.URL.Path[10:])

	path, err := authPath(w, r)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	http.ServeFile(w, r, filepath.Join(path, "snapshots", r.URL.Path[10:]))
}
