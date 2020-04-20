package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
	"os"
)

// MyEnv Variables
type MyEnv struct {
	HostName string `json:"podname,omitempty"`
	NodeName string `json:"nodename,omitempty"`
	PodIP    string `json:"podip,omitempty"`
}

//VERSION of the app
const VERSION = "v1"

var hostName string

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", GetRootEndpoint).Methods("GET")
	router.HandleFunc("/fail", GetFailEndpoint).Methods("GET")
	
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "Unknown"
	}
	fmt.Printf("%s - %s Server Start\n", time.Now().String(), hostName)
	log.Fatal(http.ListenAndServe(":80", router))
}

//LogIt helps lof
func LogIt(pipeline string, req *http.Request) {
	ua := req.Header.Get("User-Agent")
	ip := req.Header.Get("X-Forwarded-For")
	fmt.Printf("%s,%s,%s,%s\n", time.Now(), pipeline, ua, ip)
}

//GetRootEndpoint gets a Root Endpoint
func GetRootEndpoint(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Printf("%s - %s:Healthprobe Success\n", time.Now().String(), hostName)
}

//GetFailEndpoint gets a Root Endpoint
func GetFailEndpoint(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("%s - %s:Healthprobe Failed\n", time.Now().String(), hostName)
	w.WriteHeader(http.StatusNotFound)
}
