package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// mock do publisher (não usa RabbitMQ real)
type MockPublisher struct{}

func (m *MockPublisher) Publish(body []byte) error {
	return nil
}

// teste de sucesso
func TestTelemetryHandler_Success(t *testing.T) {
	mockPub := &MockPublisher{}

	reqBody := `{
		"device_id":"dev-001",
		"timestamp":"2026-03-17T15:00:00Z",
		"sensor_type":"temperature",
		"reading_type":"analog",
		"value":26.7
	}`

	req := httptest.NewRequest(http.MethodPost, "/telemetry", strings.NewReader(reqBody))
	w := httptest.NewRecorder()

	handler := telemetryHandler(mockPub)
	handler(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("esperado %d, obtido %d", http.StatusAccepted, w.Code)
	}
}

// teste de JSON inválido
func TestTelemetryHandler_InvalidJSON(t *testing.T) {
	mockPub := &MockPublisher{}

	req := httptest.NewRequest(http.MethodPost, "/telemetry", strings.NewReader("{invalid"))
	w := httptest.NewRecorder()

	handler := telemetryHandler(mockPub)
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("esperado %d, obtido %d", http.StatusBadRequest, w.Code)
	}
}

// teste de método inválido
func TestTelemetryHandler_MethodNotAllowed(t *testing.T) {
	mockPub := &MockPublisher{}

	req := httptest.NewRequest(http.MethodGet, "/telemetry", nil)
	w := httptest.NewRecorder()

	handler := telemetryHandler(mockPub)
	handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("esperado %d, obtido %d", http.StatusMethodNotAllowed, w.Code)
	}
}