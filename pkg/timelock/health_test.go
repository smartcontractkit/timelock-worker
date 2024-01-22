package timelock

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	// Initialize healthStatus to HealthStatusOK
	SetLiveStatus(HealthStatusOK)

	// Create a request to the healthz endpoint
	req := httptest.NewRequest("GET", "http://example.com/healthz", nil)
	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Call the liveHandler function
	liveHandler(rr, req)

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
	SetLiveStatus(HealthStatusError)

	// Create a new request to the healthz endpoint
	req = httptest.NewRequest("GET", "http://example.com/healthz", nil)
	// Create a new ResponseRecorder
	rr = httptest.NewRecorder()

	// Call the liveHandler function again
	liveHandler(rr, req)

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
