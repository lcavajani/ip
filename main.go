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
	IPv4Bits          int    = 32
	IPv6Bits          int    = 128
	urlDefaultPath    string = "/"
	urlInfoPath       string = "/info"
	urlPathIPv4       string = "/ip4calc"
	urlPathIPv6       string = "/ip6calc"
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
	a.Router.HandleFunc(urlDefaultPath, getRemoteIP).Methods(http.MethodGet)
	a.Router.HandleFunc(urlInfoPath, getHTTPInfo).Methods(http.MethodGet)
	a.Router.HandleFunc(urlPathIPv4, getIPCalc).Methods(http.MethodGet)
}

func (a *App) run(port string) {
	log.Fatal(http.ListenAndServe(":"+port, a.Router))
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, errorJSON{Error: message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorJSON{Error: err.Error()})
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
	var ipCalc Calculate
	var IP net.IP
	var cidr int

	// populate r.Form
	r.ParseForm()

	// ip and cidr must be provided
	if len(r.Form) != 2 {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("wrong number of arguments: %v", len(r.Form)))
	}

	// Get query parameters
	for k, v := range r.Form {
		val := v[0]
		switch k {
		case "ip":
			// ParseIP returns nil if the IP is wrong
			IP = net.ParseIP(val)
			if IP == nil {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid ip: %s", val))
				return
			}
		case "cidr":
			var err error
			cidr, err = strconv.Atoi(val)
			if (err != nil) || (cidr > IPv6Bits) {
				respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid cidr: %s", val))
				return
			}
		default:
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("wrong argument provided: %v", k))
			return
		}
	}

	switch r.URL.Path {
	case urlPathIPv4:
		if (IP.To4() == nil) || (net.CIDRMask(cidr, IPv4Bits) == nil) {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid IPv4: %s/%d", IP.String(), cidr))
			return
		}

		ipCalc = newIPv4Calc(IP.String(), cidr)
	case urlPathIPv6:
		if (IP.To16() == nil) || (net.CIDRMask(cidr, IPv6Bits) == nil) {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid IPv6: %s/%d", IP.String(), cidr))
			return
		}

		ipCalc = newIPv4Calc(IP.String(), cidr)
	}

	err := calculate(ipCalc)
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
