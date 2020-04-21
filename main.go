package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

var hostName string
var heartbeatURL string
var retCode int

func main() {
	var err error
	var hbURL string
	hostName, err = os.Hostname()
	if err != nil {
		hostName = "Unknown"
	}
	fmt.Printf("%s - %s Starting Server...\n", time.Now().String(), hostName)

	//Extract Config Variables
	retCode, hbURL = GetConfigFromArgs(os.Args)
	if retCode == -1 {
		fmt.Printf("%s:%s: - Aborting - %s\n", time.Now().String(), hostName, hbURL)
		return
	}
	heartbeatURL = hbURL
	fmt.Printf("%s - %s RETCODE...%d\t%s\n", time.Now().String(), hostName, retCode, hbURL)

	router := mux.NewRouter()
	router.HandleFunc("/", GetRootEndpoint).Methods("GET")
	router.HandleFunc("/fail", GetFailEndpoint).Methods("GET")
	router.HandleFunc("/healthprobe", GetHPEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(":80", router))
}

//GetConfigFromArgs - Read from Env
func GetConfigFromArgs(args []string) (int, string) {
	var err error
	if len(os.Args) != 3 {
		fmt.Println("No Arguments")
		fmt.Println("Usage is :  app [retcode] [activeendpoint]")
		fmt.Println("app [200/400] [http://127.0.0.1/healthprobe]")
		return -1, ""
	}
	rCode, err := strconv.Atoi(os.Args[1])
	if err != nil {
		return -1, err.Error()
	}
	//Check of rCode Range
	hbURL := os.Args[2]
	//fmt.Printf("%s:%s:Args:%d %s\n", time.Now().String(), hostName, rCode, hbURL)

	return rCode, hbURL
}

//LogIt helps lof
func LogIt(pipeline string, req *http.Request) {
	ua := req.Header.Get("User-Agent")
	ip := req.Header.Get("X-Forwarded-For")
	fmt.Printf("%s,%s,%s,%s\n", time.Now(), pipeline, ua, ip)
}

//GetHPEndpoint gets a HP Endpoint
func GetHPEndpoint(w http.ResponseWriter, req *http.Request) {
	/*
		if retCode=404 and HeartBeat URL is specified,
		check HeartBeat URL for status
		if HeartBeat URL is not 200, return 404
	*/
	var rCode int
	//fmt.Printf("%s - %s:GetHPEndpoint  %d\t%s\n", time.Now().String(), hostName, retCode, heartbeatURL)
	if retCode == 200 && strings.Index(heartbeatURL, "localhost") > 0 {
		//You are the leader nothing to do
		rCode = 200
	} else if retCode == 400 {
		//Check the leader
		rCode = GetHeartBeat()
		if rCode != 200 {
			//Make yourself a Leader
			fmt.Printf("%s - %s:Transforming to Leader\n", time.Now().String(), hostName)
			rCode = 200
		} else {
			rCode = 400
		}
	} else {
		rCode = 200
	}
	fmt.Printf("%s - %s:HeartBeat %d\n", time.Now().String(), hostName, rCode)
	w.WriteHeader(rCode)

}

//GetRootEndpoint gets a Root Endpoint
func GetRootEndpoint(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Printf("%s - %s:ROOT Success\n", time.Now().String(), hostName)
}

//GetFailEndpoint gets a Root Endpoint
func GetFailEndpoint(w http.ResponseWriter, req *http.Request) {
	/*
		if retCode=404 and HeartBeat URL is specified,
		check HeartBeat URL for status
		if HeartBeat URL is not 200, return 404
	*/
	var rCode int
	if retCode == 200 && strings.Index(heartbeatURL, "localhost") > 0 {
		//You are the leader nothing to do
		rCode = 200
	} else if retCode == 400 {
		//Check the leader
		rCode = GetHeartBeat()
		if rCode != 200 {
			//Make yourself a Leader
			rCode = 200
		} else {
			rCode = 400
		}
	}
	fmt.Printf("%s - %s:Healthprobe Status %d\n", time.Now().String(), hostName, rCode)
	w.WriteHeader(rCode)
}

//GetHeartBeat gets a Root Endpoint
func GetHeartBeat() int {
	var retStatus int

	res, err := http.Get(heartbeatURL)
	if err != nil {
		retStatus = 404
	} else {
		retStatus = res.StatusCode
		res.Close = true
		res.Body.Close()
	}


	//fmt.Printf("%s - %s:GetHeartBeat  %d\n", time.Now().String(), hostName, res.StatusCode)

	return retStatus
}
