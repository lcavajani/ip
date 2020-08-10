package main

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

var a *App

// TestNewApp uses a different way to check the returned type using reflect
func TestNewApp(t *testing.T) {
	app := newApp()
	actual := reflect.TypeOf(*app)
	expected := reflect.TypeOf((*App)(nil)).Elem()
	checkResultVSExpected(t, "returned struct is not from expected type", actual, expected)
}

func TestRespondWithError(t *testing.T) {
	testErrorMessage := "Test error message"
	req, err := http.NewRequest("GET", "/dummy", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// it is not possible to unmarshal a function so we should get an internal error
		respondWithError(w, http.StatusForbidden, testErrorMessage)
	})
	handler.ServeHTTP(rr, req)

	var resultBody map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &resultBody)
	if err != nil {
		t.Errorf("error unmarshalling")
	}

	checkResultVSExpected(t, "wrong error message", resultBody["error"], testErrorMessage)
	checkResultVSExpected(t, "wrong return code", rr.Code, http.StatusForbidden)
}

func TestRespondWithJSON(t *testing.T) {
	req, err := http.NewRequest("GET", "/dummy", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// it is not possible to unmarshal a function so we should get an internal error
		respondWithJSON(w, http.StatusOK, func() {})
	})
	handler.ServeHTTP(rr, req)

	checkResultVSExpected(t, "wrong return code", rr.Code, http.StatusInternalServerError)
}

func TestGetRemoteIP(t *testing.T) {
	var tests = []string{"10.0.0.1", "1.1.1.1", "192.168.0.1"}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/", nil)
		if err != nil {
			t.Fatal(err)
		}

		// Set remote address in request IP:PORT
		port := strconv.Itoa(rand.Intn(65535))
		req.RemoteAddr = net.JoinHostPort(tt, port)
		rr := executeRequest(req)

		checkResultVSExpected(t, "handler returned wrong remote IP", rr.Body.String(), tt)
	}
}

func TestGetHTTPInfo(t *testing.T) {
	type expected struct {
		Header     http.Header `json:"header"`
		Host       string      `json:"host"`
		RemoteAddr string      `json:"remoteAddr"`
	}

	var tests = []expected{
		{
			map[string][]string{
				"Accept-Encoding": {"gzip, deflate"},
				"User-Agent":      {"test"},
			},
			"192.168.0.1", "192.168.0.254",
		},
	}

	for _, tt := range tests {
		req, err := http.NewRequest("GET", "/info", nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header = tt.Header
		req.RemoteAddr = tt.RemoteAddr
		req.Host = tt.Host

		rr := executeRequest(req)

		var result expected
		err = json.Unmarshal(rr.Body.Bytes(), &result)
		if err != nil {
			t.Errorf("error unmarshalling")
		}

		switch {
		case !reflect.DeepEqual(result.Header, tt.Header):
			t.Errorf("wrong headers returned in body\ngot: %+v\nwant: %+v", tt.Header, result.Header)
		case result.Host != tt.Host:
			t.Errorf("returned host is wrong\ngot: %v\nwant:%v", result.Host, tt.Host)
		case result.RemoteAddr != tt.RemoteAddr:
			t.Errorf("returned remote addr is wrong\ngot: %v\nwant:%v", result.RemoteAddr, tt.RemoteAddr)
		}
	}

}

func TestGetIPCalc(t *testing.T) {
	var tests = []struct {
		ip      string
		cidr    string
		retCode int
	}{
		{"1.1.1.1", "1", http.StatusOK},
		{"192.168.0.1", "24", http.StatusOK},
		{"192.168.0.1", "31", http.StatusOK},
		{"192.168.0.1", "32", http.StatusOK},
		{"", "", http.StatusBadRequest},
		{"", "24", http.StatusBadRequest},
		{"192.168.0.1", "", http.StatusBadRequest},
		{"192.168.0.1111", "24", http.StatusBadRequest},
		{"192.168.0.1", "24444", http.StatusBadRequest},
		{"192.168.0.1", "-----", http.StatusBadRequest},
		{"e80::", "10", http.StatusBadRequest},
		{"192.168.0.1", "128", http.StatusBadRequest},
	}

	for _, tt := range tests {
		form := url.Values{}
		form.Add("ip", tt.ip)
		form.Add("cidr", tt.cidr)
		req, err := http.NewRequest("GET", "/ip4calc", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.PostForm = form
		rr := executeRequest(req)
		checkResultVSExpected(t, "wrong return code", rr.Code, tt.retCode)
	}

	// Test with wrong params
	badParams := []map[string]string{
		{"na": "192.168.0.1", "ip": "192.168.0.1"},
		{"na": "192.168.0.1", "ip": "192.168.0.1", "cidr": "24"},
	}
	for _, params := range badParams {
		form := url.Values{}
		for k, v := range params {
			form.Add(k, v)
		}

		req, err := http.NewRequest("GET", "/ip4calc", strings.NewReader(form.Encode()))
		if err != nil {
			t.Fatal(err)
		}
		req.PostForm = form
		rr := executeRequest(req)
		checkResultVSExpected(t, "wrong return code", rr.Code, http.StatusBadRequest)
	}

}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResultVSExpected(t *testing.T, message string, result, expected interface{}) {
	if result != expected {
		t.Errorf("%s\ngot:  %v\nwant: %v", message, result, expected)
	}
}

func TestMain(m *testing.M) {
	a = newApp()
	a.initialize()
	code := m.Run()
	os.Exit(code)
}
