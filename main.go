package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const (
	headerContentType string = "Content-Type"
	headerAppJson     string = "application/json"
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
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ipCalcError{Error: err.Error()})
		return
	}

	w.Header().Set(headerContentType, headerAppJson)
	w.WriteHeader(code)
	w.Write(response)
}

// getRemoteIp will return the remote address of the client
func getRemoteIp(w http.ResponseWriter, r *http.Request) {
	remoteAddr := strings.Split(r.RemoteAddr, ":")
	io.WriteString(w, remoteAddr[0])
	return
}

// getHttpInfo will return the request headers and the address of the host and the client
func getHttpInfo(w http.ResponseWriter, r *http.Request) {
	var httpInfo = struct {
		Header     http.Header `json:"header"`
		Host       string      `json:"host"`
		RemoteAddr string      `json:"remoteAddr"`
	}{
		Header:     r.Header,
		Host:       r.Host,
		RemoteAddr: r.RemoteAddr,
	}

	respondWithJSON(w, http.StatusOK, httpInfo)
	return
}

func getIpCalc(w http.ResponseWriter, r *http.Request) {
	ipCalc := NewIpCalc()

	// populate r.Form
	r.ParseForm()

	// ip and cidr must be provided
	if len(r.Form) != 2 {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("wrong number of arguments: %v", len(r.Form)))
		return
	}

	// Get query parameters
	for k, v := range r.Form {
		val := v[0]
		switch k {
		case "ip":
			// ParseIP returns nil if the IP is wrong
			if ip := net.ParseIP(val); ip == nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid ip: %v", val))
				return
			}
			ipCalc.Address = val
		case "cidr":
			valInt, _ := strconv.Atoi(val)
			// make sure val is set to avoid defaulting on /0 from default struct val
			if (net.CIDRMask(valInt, 32) == nil) || (val == "") {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid cidr: %d", valInt))
				return
			}
			ipCalc.Cidr = valInt
		default:
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("wrong argument provided: %v", k))
			return
		}
	}

	err := ipCalc.getNetworkInfo()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, ipCalc)
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
