package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

type App struct {
	Router *mux.Router
}

func NewApp() *App {
	return &App{}
}

func (a *App) Initialize() {
	a.Router = mux.NewRouter().StrictSlash(false)
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/", getRemoteIp).Methods("GET")
	a.Router.HandleFunc("/info", getHttpInfo).Methods("GET")
	a.Router.HandleFunc("/ipcalc", getIpCalc).Methods("GET")
}

func (a *App) Run(port string) {
	log.Fatal(http.ListenAndServe(":"+port, a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, ipCalcError{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// getRemoteIp will return the remote address of the client
func getRemoteIp(w http.ResponseWriter, r *http.Request) {
	remoteAddr := strings.Split(r.RemoteAddr, ":")
	fmt.Fprint(w, remoteAddr[0])
	return
}

// getHttpInfo will return the request headers and the address of the host and the client
func getHttpInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var httpInfo = struct {
		Headers    *http.Header `json:"headers"`
		Host       *string      `json:"host"`
		RemoteAddr *string      `json:"remoteAddr"`
	}{
		Headers:    &r.Header,
		Host:       &r.Host,
		RemoteAddr: &r.RemoteAddr,
	}

	respondWithJSON(w, http.StatusOK, httpInfo)
	return
}

func main() {
	a := NewApp()
	a.Initialize()

	// Get env var for http port
	http_port := "8000"
	if hp := os.Getenv("HTTP_PORT"); hp != "" {
		http_port = hp
	}

	a.Run(http_port)
}
