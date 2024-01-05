package timelock

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

// HealthStatus represents the health status enum
type HealthStatus int

const (
	HealthStatusOK HealthStatus = iota
	HealthStatusError
)

var healthStatus atomic.Value

// SetHealthStatus sets the health status
func SetHealthStatus(status HealthStatus) {
	healthStatus.Store(status)
}

// GetHealthStatus gets the current health status
func GetHealthStatus() HealthStatus {
	return healthStatus.Load().(HealthStatus)
}

// Respond to liveness probe based on health status.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	status := GetHealthStatus()
	if status == HealthStatusOK {
		w.Write([]byte("OK"))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error"))
	}
}

// Starts a http server, serving the healthz endpoint.
func StartHTTPHealthServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthHandler)

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
