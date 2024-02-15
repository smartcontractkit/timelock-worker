package timelock

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
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
	respond(status, w)
}

// Respond to readiness probe based on ready status.
func readyHandler(w http.ResponseWriter, r *http.Request) {
	status := GetReadyStatus()
	respond(status, w)
}

func respond(status HealthStatus, w http.ResponseWriter) {
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
func StartHTTPHealthServer(l *zerolog.Logger) {
	startServer(l, "health", ":8080", func(mux *http.ServeMux) {
		mux.HandleFunc("/healthz", liveHandler)
		mux.HandleFunc("/ready", readyHandler)
	})
}

func StartMetricsServer(l *zerolog.Logger) {
	startServer(l, "metrics", ":2021", func(mux *http.ServeMux) {
		mux.Handle("/metrics", promhttp.Handler())
	})
}

func startServer(l *zerolog.Logger, name, addr string, opts ...func(*http.ServeMux)) {
	mux := http.NewServeMux()
	for _, opt := range opts {
		opt(mux)
	}

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,  // Set your desired read timeout
		WriteTimeout: 10 * time.Second, // Set your desired write timeout
		IdleTimeout:  15 * time.Second, // Set your desired idle timeout
	}

	l.Info().Msgf("%s server listening on %s", name, addr)
	if err := server.ListenAndServe(); err != nil {
		l.Error().Msgf("%s server stopped: %s", name, err)
	}
}
