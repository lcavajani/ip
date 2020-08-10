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

	"github.com/gorilla/mux"
)

const (
	headerContentType string = "Content-Type"
	headerAppJSON     string = "application/json"
)

// A App defines the router
type App struct {
	Router *mux.Router
}

func newApp() *App {
	return &App{}
}

func (a *App) initialize() {
	a.Router = mux.NewRouter().StrictSlash(false)
	a.initializeRoutes()
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/", getRemoteIP).Methods(http.MethodGet)
	a.Router.HandleFunc("/info", getHTTPInfo).Methods(http.MethodGet)
	a.Router.HandleFunc("/ipcalc", getIPCalc).Methods(http.MethodGet)
}

func (a *App) run(port string) {
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

	w.Header().Set(headerContentType, headerAppJSON)
	w.WriteHeader(code)
	w.Write(response)
}

// getRemoteIP will return the remote address of the client
func getRemoteIP(w http.ResponseWriter, r *http.Request) {
	remoteAddr, _, _ := net.SplitHostPort(r.RemoteAddr)
	io.WriteString(w, remoteAddr)
	return
}

// getHTTPInfo will return the request headers and the address of the host and the client
func getHTTPInfo(w http.ResponseWriter, r *http.Request) {
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

func getIPCalc(w http.ResponseWriter, r *http.Request) {
	ipCalc := newIPCalc()

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
	a := newApp()
	a.initialize()

	// Get env var for http port
	httpPort := "8000"
	if hp := os.Getenv("HTTP_PORT"); hp != "" {
		httpPort = hp
	}

	a.run(httpPort)
}
