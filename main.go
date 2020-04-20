package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"time"
)

// MyEnv Variables
type MyEnv struct {
	HostName string `json:"podname,omitempty"`
	NodeName string `json:"nodename,omitempty"`
	PodIP    string `json:"podip,omitempty"`
}

//VERSION of the app
const VERSION = "v1"

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", GetRootEndpoint).Methods("GET")
	router.HandleFunc("/fail", GetFailEndpoint).Methods("GET")
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

	var thisenv MyEnv
	thisenv.HostName = "HOST"
	thisenv.NodeName = "NODE"
	thisenv.PodIP = "IP"
	bytesRepresentation, _ := json.Marshal(thisenv)
	fmt.Println(string(bytesRepresentation))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytesRepresentation)
}

//GetFailEndpoint gets a Root Endpoint
func GetFailEndpoint(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
