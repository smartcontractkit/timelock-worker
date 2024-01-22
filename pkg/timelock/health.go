package timelock

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthStatus represents the health status enum.
type HealthStatus int

const (
	HealthStatusOK HealthStatus = iota
	HealthStatusError
)

var liveStatus atomic.Value
var readyStatus atomic.Value

func SetLiveStatus(status HealthStatus) {
	liveStatus.Store(status)
}

func GetLiveStatus() HealthStatus {
	return liveStatus.Load().(HealthStatus)
}

func SetReadyStatus(status HealthStatus) {
	readyStatus.Store(status)
}

func GetReadyStatus() HealthStatus {
	return readyStatus.Load().(HealthStatus)
}

func liveHandler(w http.ResponseWriter, r *http.Request) {
	status := GetLiveStatus()
	respond(status, w, r)
}

// Respond to readiness probe based on ready status.
func readyHandler(w http.ResponseWriter, r *http.Request) {
	status := GetReadyStatus()
	respond(status, w, r)
}

func respond(status HealthStatus, w http.ResponseWriter, r *http.Request) {
	var err error
	if status == HealthStatusOK {
		_, err = w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		_, err = w.Write([]byte("Error"))
	}

	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

// Starts a http server, serving the healthz endpoint.
func StartHTTPHealthServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", liveHandler)
	mux.HandleFunc("/ready", readyHandler)

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,  // Set your desired read timeout
		WriteTimeout: 10 * time.Second, // Set your desired write timeout
		IdleTimeout:  15 * time.Second, // Set your desired idle timeout
	}

	fmt.Println("Server listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		fmt.Println(err)
	}
}
