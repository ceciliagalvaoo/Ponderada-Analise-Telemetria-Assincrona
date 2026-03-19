package main

import (
	"errors"
	"testing"
	"time"
)

type MockRepository struct {
	Called      bool
	DeviceID    string
	Timestamp   time.Time
	SensorType  string
	ReadingType string
	Value       float64
	Err         error
}

func (m *MockRepository) InsertTelemetry(deviceID string, timestamp time.Time, sensorType, readingType string, value float64) error {
	m.Called = true
	m.DeviceID = deviceID
	m.Timestamp = timestamp
	m.SensorType = sensorType
	m.ReadingType = readingType
	m.Value = value
	return m.Err
}

func TestProcessMessage_Success(t *testing.T) {
	mockRepo := &MockRepository{}

	body := []byte(`{
		"device_id":"dev-001",
		"timestamp":"2026-03-17T15:00:00Z",
		"sensor_type":"temperature",
		"reading_type":"analog",
		"value":26.7
	}`)

	err := processMessage(mockRepo, body)
	if err != nil {
		t.Fatalf("nao esperava erro, mas recebeu: %v", err)
	}

	if !mockRepo.Called {
		t.Fatal("esperava que InsertTelemetry fosse chamado")
	}

	if mockRepo.DeviceID != "dev-001" {
		t.Errorf("device_id esperado 'dev-001', obtido '%s'", mockRepo.DeviceID)
	}

	if mockRepo.SensorType != "temperature" {
		t.Errorf("sensor_type esperado 'temperature', obtido '%s'", mockRepo.SensorType)
	}

	if mockRepo.ReadingType != "analog" {
		t.Errorf("reading_type esperado 'analog', obtido '%s'", mockRepo.ReadingType)
	}

	if mockRepo.Value != 26.7 {
		t.Errorf("value esperado 26.7, obtido %f", mockRepo.Value)
	}
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	mockRepo := &MockRepository{}

	body := []byte(`{invalid json`)

	err := processMessage(mockRepo, body)
	if err == nil {
		t.Fatal("esperava erro para JSON invalido")
	}

	if mockRepo.Called {
		t.Fatal("nao esperava que InsertTelemetry fosse chamado")
	}
}

func TestProcessMessage_InvalidTimestamp(t *testing.T) {
	mockRepo := &MockRepository{}

	body := []byte(`{
		"device_id":"dev-001",
		"timestamp":"data-invalida",
		"sensor_type":"temperature",
		"reading_type":"analog",
		"value":26.7
	}`)

	err := processMessage(mockRepo, body)
	if err == nil {
		t.Fatal("esperava erro para timestamp invalido")
	}

	if mockRepo.Called {
		t.Fatal("nao esperava que InsertTelemetry fosse chamado")
	}
}

func TestProcessMessage_DBError(t *testing.T) {
	mockRepo := &MockRepository{
		Err: errors.New("erro no banco"),
	}

	body := []byte(`{
		"device_id":"dev-001",
		"timestamp":"2026-03-17T15:00:00Z",
		"sensor_type":"temperature",
		"reading_type":"analog",
		"value":26.7
	}`)

	err := processMessage(mockRepo, body)
	if err == nil {
		t.Fatal("esperava erro vindo do banco")
	}

	if !mockRepo.Called {
		t.Fatal("esperava que InsertTelemetry fosse chamado")
	}
}
