package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go-sf-newsaggr-36a4-1/pkg/storage"

	"github.com/gorilla/mux"
)

type API struct {
	db *storage.DB
	r  *mux.Router
	hp string //home path
}

// constructor of API
func New(db *storage.DB, homepath string) *API {
	a := API{db: db, r: mux.NewRouter(), hp: homepath}
	a.endpoints()
	return &a
}

// returns router for HTTP Server
func (api *API) Router() *mux.Router {
	return api.r
}

// register endpoints
func (api *API) endpoints() {
	// select last <n> news
	api.r.HandleFunc("/news/{n}", api.posts).Methods(http.MethodGet, http.MethodOptions)

	// html web server
	api.r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir(api.hp+"/webapp"))))

}

// GET /news/{n} - returns <n> last posts
func (api *API) posts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		return
	}
	s := mux.Vars(r)["n"]
	n, _ := strconv.Atoi(s)
	news, err := api.db.News(n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(news)
}
