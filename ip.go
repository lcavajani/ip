package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

type IpCalc struct {
	Address        string `json:"address,omitempty"`
	Cidr           string `json:"cidr,omitempty"`
	Netmask        string `json:"netmask,omitempty"`
	NetworkCidr    string `json:"networkCidr,omitempty"`
	Network        string `json:"network,omitempty"`
	HostMin        string `json:"hostMin,omitempty"`
	HostMax        string `json:"hostMax,omitempty"`
	Broadcast      string `json:"broadcast,omitempty"`
	HostsAvailable string `json:"hostsAvailable,omitempty"`
	HostsTotal     string `json:"hostsTotal,omitempty"`
	Error          string `json:"error,omitempty"`
}

type HttpInfo struct {
	Headers    *http.Header `json:"headers"`
	Host       *string      `json:"host"`
	RemoteAddr *string      `json:"remoteAddr"`
}

func main() {
	// Get env var for http port
	http_port := "8000"
	if hp := os.Getenv("HTTP_PORT"); hp != "" {
		http_port = hp
	}

	// Create router
	router := mux.NewRouter().StrictSlash(false)

	// Create routers
	router.HandleFunc("/", getRemoteIp).Methods("GET")
	router.HandleFunc("/info", getHttpInfo).Methods("GET")
	router.HandleFunc("/ipcalc", getIpCalc).Methods("GET")

	// Start HTTP server
	log.Fatal(http.ListenAndServe(":"+http_port, router))
}

func getRemoteIp(w http.ResponseWriter, r *http.Request) {
	remoteAddr := strings.Split(r.RemoteAddr, ":")
	fmt.Fprintf(w, remoteAddr[0])
}

func getHttpInfo(w http.ResponseWriter, r *http.Request) {
	httpInfo := HttpInfo{}
	w.Header().Set("Content-Type", "application/json")

	httpInfo.Headers = &r.Header
	httpInfo.Host = &r.Host
	httpInfo.RemoteAddr = &r.RemoteAddr

	json.NewEncoder(w).Encode(httpInfo)
	return
}

func getIpCalc(w http.ResponseWriter, r *http.Request) {
	var ips []string
	ipCalc := IpCalc{}
	w.Header().Set("Content-Type", "application/json")

	// populate r.Form
	r.ParseForm()

	// Test number of arguments
	if len(r.Form) != 2 {
		ipCalc.Error = fmt.Sprintf("wrong number of arguments: %v", len(r.Form))
		json.NewEncoder(w).Encode(ipCalc)
		return
	}

	// Get query parameters
	for k, v := range r.Form {
		val := strings.Join(v, "")
		switch k {
		case "ip":
			// ParseIP returns nil if the IP is wrong
			if ip := net.ParseIP(val); ip == nil {
				ipCalc.Error = fmt.Sprintf("invalid ip: %v", val)
				json.NewEncoder(w).Encode(ipCalc)
				return
			}
			ipCalc.Address = val
		case "cidr":
			valInt, _ := strconv.Atoi(val)
			// Since inc function is really not optimal, avoid bigger calculation
			if (valInt < 8) || (net.CIDRMask(valInt, 32) == nil) {
				ipCalc.Error = fmt.Sprintf("invalid or too low (<8) cidr: %v", val)
				json.NewEncoder(w).Encode(ipCalc)
				return
			}
			ipCalc.Cidr = val
		default:
			ipCalc.Error = fmt.Sprintf("wrong argument provided: %v", k)
			json.NewEncoder(w).Encode(ipCalc)
			return
		}
	}

	ip, ipNet, err := net.ParseCIDR(ipCalc.Address + "/" + ipCalc.Cidr)
	if err != nil {
		ipCalc.Error = fmt.Sprintf("something went wrong during cidr parsing: %v", err)
		json.NewEncoder(w).Encode(ipCalc)
		return
	}

	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}

	// Available for all cidr
	ipCalc.Netmask = net.IP(ipNet.Mask).String()
	ipCalc.HostsTotal = strconv.Itoa(len(ips))

	switch len(ips) {
	// Case of a /32
	case 1:
		ipCalc.HostMin = ips[0]
		ipCalc.HostMax = ips[0]
		ipCalc.HostsAvailable = "1"
	// Case of a /31
	case 2:
		ipCalc.HostMin = ips[0]
		ipCalc.HostMax = ips[1]
		ipCalc.NetworkCidr = ipNet.String()
		ipCalc.Network = strings.Split(ipCalc.NetworkCidr, "/")[0]
		ipCalc.HostsAvailable = "2"
	// Anything else
	default:
		ipCalc.HostMin = ips[1]
		ipCalc.HostMax = ips[len(ips)-2]
		ipCalc.Broadcast = ips[len(ips)-1]
		ipCalc.NetworkCidr = ipNet.String()
		ipCalc.HostsAvailable = strconv.Itoa(len(ips) - 2)
	}

	json.NewEncoder(w).Encode(ipCalc)
	return
}

func inc(ip net.IP) {
	for l := len(ip) - 1; l >= 0; l-- {
		ip[l]++
		if ip[l] > 0 {
			break
		}
	}
}
