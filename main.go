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

type stop struct {
	error
}

//HTTPhandler for Retry
type HTTPhandler func() error

//TCPhandler for Retry
type TCPhandler func() error

// MyEnv Variables
type MyEnv struct {
	HostName string `json:"podname,omitempty"`
	NodeName string `json:"nodename,omitempty"`
	PodIP    string `json:"podip,omitempty"`
}

//VERSION of the app
const VERSION = "v1"
const TCP_DIAL_TIMEOUT_SECS = 10
const NETWORK_RETRIES = 3
const NETWORK_RETRY_BACKOFF_SECS = 2

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
		log.Printf("%s - %s Received signal %s. Shutting down Server %s...\n", GetTimeAsString(), role, s, hostName)
		AppCleanup()
		os.Exit(1)
	}()

	log.Printf("%s - %s Starting %s Server...\n", GetTimeAsString(), role, hostName)
	log.Printf("tcpprobe:%s agentport:%s\n", tcpprobe, agentport)

	router := mux.NewRouter()
	router.HandleFunc("/", GetRootEndpoint).Methods("GET")
	router.HandleFunc("/healthprobe", GetHPEndpoint).Methods("GET")
	log.Fatal(http.ListenAndServe(agentport, router))
}

//AppCleanup on Kill
func AppCleanup() {
	log.Printf("%s - %s  %s Server is Exiting...\n", GetTimeAsString(), role, hostName)
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

	err := Retry(NETWORK_RETRIES, time.Duration(time.Second*NETWORK_RETRY_BACKOFF_SECS), GetHeartBeatTCP)
	if err != nil {
		tcpStatus = false
	} else {
		tcpStatus = true
	}
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
			log.Printf("%s - Secondary is Active:%s\n", GetTimeAsString(), hostName)
			rCode = 200
		} else {
			//Primary is healthy , fake Secondary is not healthy
			rCode = 404
		}
	}
	log.Printf("%s - HeartBeat:%s %d\n", GetTimeAsString(), hostName, rCode)
	w.WriteHeader(rCode)
}

//GetRootEndpoint gets a Root Endpoint
func GetRootEndpoint(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{%v: %v}\n", hostName, http.StatusOK)
	log.Printf("%s - %s:%s Success\n", GetTimeAsString(), role, hostName)
}

//GetHeartBeatHTTP connects to tcpprobe and enquires status
func GetHeartBeatHTTP() error {
	res, err := http.Get(heartbeatURL)
	if err != nil {
		return err
	}
	res.Close = true
	res.Body.Close()
	return nil
}

//GetHeartBeatTCP gets a Root Endpoint
func GetHeartBeatTCP() error {
	conn, err := net.DialTimeout("tcp", tcpprobe, time.Duration(TCP_DIAL_TIMEOUT_SECS)*time.Second)
	if err != nil {
		log.Printf("%s - %s:tcpprobe conn error: %s\n", GetTimeAsString(), hostName, tcpprobe)
		return err
	}

	defer conn.Close()
	if err, ok := err.(*net.OpError); ok && err.Timeout() {
		log.Printf("%s - %s:TCP Timeout: %s\n", GetTimeAsString(), hostName, tcpprobe)
		log.Printf("Timeout error: %s\n", err)
		return err
	}

	return nil
}

//Retry retries the specific function until Stop or attempts
func Retry(attempts int, sleep time.Duration, fn HTTPhandler) error {
	log.Printf("Attempt:%d\n", attempts)
	if err := fn(); err != nil {
		if s, ok := err.(stop); ok {
			// Return the original error for later checking
			return s.error
		}
		if attempts--; attempts > 0 {
			time.Sleep(sleep)

			return Retry(attempts, 2*sleep, fn)
		}
		return err
	}
	return nil
}

//GetTimeAsString for logging
func GetTimeAsString() string {
	currentTime := time.Now()
	return currentTime.Format("2020-01-01 01:01:01")
}
