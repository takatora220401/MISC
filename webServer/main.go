package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		panic(err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler:   mux, 
	}

	mux.HandleFunc("/hostname", Hostname)
	mux.HandleFunc("/headers", Headers)
	mux.HandleFunc("/cookies", Cookies)
	mux.HandleFunc("/srcip", TCPpeerIP)
	mux.HandleFunc("/responsetime", ResponseTime)

	err = server.ListenAndServeTLS("", "")
	if err != nil {
		panic(err)
	}
}

func Hostname(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	// Start writing response
	fmt.Fprintf(w, "Hostname: %s\n", hostname)
}

func Headers(w http.ResponseWriter, r *http.Request) {
	for name, values := range r.Header {
		for _, v := range values {
			fmt.Fprintf(w, "%s: %s\n", name, v)
		}
	}
}

func Cookies(w http.ResponseWriter, r *http.Request) {
	cookies := r.Cookies()
	if len(cookies) == 0 {
		fmt.Fprintf(w, "No cookies\n")
	} else {
		for _, c := range cookies {
			fmt.Fprintf(w, "%s=%s\n", c.Name, c.Value)
		}
	}
}

func TCPpeerIP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Client IP: %s\n", r.RemoteAddr)
}

func ResponseTime(w http.ResponseWriter, r *http.Request) {
	responseTime := time.Now()
	fmt.Fprintf(w, "%s\n", responseTime.Format(time.RFC3339Nano))
}