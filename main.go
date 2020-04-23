package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"flag"

	"strings"

	"github.com/gorilla/mux"
	"syscall"
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
var role string
var tcpprobe string

//var httpprobe string
var agentport string

func init() {
	flag.StringVar(&role, "role", "primary", "specify the role: primary/secondary")
	flag.StringVar(&tcpprobe, "tcpprobe", "127.0.0.1:8080", "specify tcp probe: x.x.x.x:port")
	//flag.StringVar(&httpprobe, "httpprobe", "http://localhost:8080/healthprobe", "specify http endpoint for primary agent: http://x.x.x.x:port/healthprobe")
	flag.StringVar(&agentport, "agentport", "8080", "specify the port for agent to run")
	flag.Parse()
	//Convert to lowercase....
	role = strings.ToLower(role)
	agentport = fmt.Sprintf(":%s", agentport)
}

func main() {
	var err error

	hostName, err = os.Hostname()
	if err != nil {
		hostName = "UnknownHost"
	}
	// setup signal catching
	sigs := make(chan os.Signal, 1)
	// catch SIGQUIT

	signal.Notify(sigs,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	// method invoked upon seeing signal
	go func() {
		s := <-sigs
		log.Printf("%s - %s Received signal %s. Shutting down Server %s...\n", time.Now().String(), role, s, hostName)
		AppCleanup()
		os.Exit(1)
	}()

	log.Printf("%s - %s Starting %s Server...\n", time.Now().String(), role, hostName)
	log.Printf("tcpprobe:%s agentport:%s\n", tcpprobe, agentport)

	router := mux.NewRouter()
	router.HandleFunc("/", GetRootEndpoint).Methods("GET")
	router.HandleFunc("/healthprobe", GetHPEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(agentport, router))
}

//AppCleanup on Kill
func AppCleanup() {
	log.Printf("%s - %s  %s Server is Exiting...\n", time.Now().String(), role, hostName)
}

//GetHPEndpoint gets a HP Endpoint
func GetHPEndpoint(w http.ResponseWriter, req *http.Request) {
	/*
		if role = primary
		{
			check MQtcp Connection
			if mq is up
				return HTTP 200 to LB
			else
				return HTTP 404 to LB
		} else {//Secondary
			Check Primary for MQ TCP Health
			if (!MQtcp)
				return HTTP 200
		}
	*/
	var tcpStatus bool
	var rCode int

	tcpStatus = GetHeartBeatTCP(tcpprobe, 10)
	if role == "primary" {
		//This block is executed on the Primary Server
		//Check MQ
		if tcpStatus == false {
			//Fail yourself, return 404
			rCode = 404
		} else {
			//All is healthy and well
			rCode = 200
		}
	} else { //Not Primary

		//Check Primary MQ Status
		if tcpStatus == false {
			//Failover to Secondary, return 200
			log.Printf("%s - Failingover to %s\n", time.Now().String(), hostName)
			rCode = 200
		} else {
			//Primary is healthy , fake Secondary is not healthy
			rCode = 404
		}
	}
	log.Printf("%s - %s:HeartBeat %d\n", time.Now().String(), hostName, rCode)
	w.WriteHeader(rCode)
}

//GetRootEndpoint gets a Root Endpoint
func GetRootEndpoint(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{%v: %v}\n", hostName, http.StatusOK)
	log.Printf("%s - %s:%s Success\n", time.Now().String(), role, hostName)
}

//GetHeartBeatHTTP gets a Root Endpoint
func GetHeartBeatHTTP() int {
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

//GetHeartBeatTCP gets a Root Endpoint
func GetHeartBeatTCP(host string, timeoutSecs int) bool {
	conn, err := net.DialTimeout("tcp", host, time.Duration(timeoutSecs)*time.Second)
	if err != nil {
		log.Printf("%s - %s:tcpprobe conn error: %s\n", time.Now().String(), hostName, host)
		return false
	}

	defer conn.Close()
	if err, ok := err.(*net.OpError); ok && err.Timeout() {
		log.Printf("%s - %s:TCP Timeout: %s\n", time.Now().String(), hostName, host)
		log.Printf("Timeout error: %s\n", err)
		return false
	}
	if err != nil {
		// Log or report the error here
		log.Printf("Error: %s\n", err)
		return false
	}
	return true
}
