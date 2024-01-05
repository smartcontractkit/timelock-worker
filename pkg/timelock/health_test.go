package timelock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Initialize healthStatus to HealthStatusOK
	SetHealthStatus(HealthStatusOK)

	// Create a request to the healthz endpoint
	req := httptest.NewRequest("GET", "http://example.com/healthz", nil)
	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the healthHandler function
	healthHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	expectedBody := "OK"
	if body := rr.Body.String(); body != expectedBody {
		t.Errorf("Handler returned unexpected body: got %v want %v", body, expectedBody)
	}

	// Change healthStatus to HealthStatusError
	SetHealthStatus(HealthStatusError)

	// Create a new request to the healthz endpoint
	req = httptest.NewRequest("GET", "http://example.com/healthz", nil)
	// Create a new ResponseRecorder
	rr = httptest.NewRecorder()

	// Call the healthHandler function again
	healthHandler(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	// Check the response body
	expectedBody = "Error"
	if body := rr.Body.String(); body != expectedBody {
		t.Errorf("Handler returned unexpected body: got %v want %v", body, expectedBody)
	}
}

func TestStartHTTPHealthServer(t *testing.T) {
	// This test just checks if the function starts without errors.
	// It doesn't perform in-depth testing of the HTTP server.

	go StartHTTPHealthServer()

	// Allow some time for the server to start
	time.Sleep(100 * time.Millisecond)

	// Create a request to the healthz endpoint
	req := httptest.NewRequest("GET", "http://localhost:8080/healthz", nil)
	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Send the request to the server
	http.DefaultServeMux.ServeHTTP(rr, req)

	// Check the response status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
