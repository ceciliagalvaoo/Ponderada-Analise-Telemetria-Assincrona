package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

func validateTelemetryPayload(t Telemetry) error {
	if strings.TrimSpace(t.DeviceID) == "" {
		return ErrInvalidDeviceID
	}

	if strings.TrimSpace(t.Timestamp) == "" {
		return ErrInvalidTimestamp
	}

	if _, err := time.Parse(time.RFC3339, t.Timestamp); err != nil {
		return ErrInvalidTimestamp
	}

	if strings.TrimSpace(t.SensorType) == "" {
		return ErrInvalidSensorType
	}

	readingType := strings.ToLower(strings.TrimSpace(t.ReadingType))
	if readingType != "analog" && readingType != "discrete" {
		return ErrInvalidReadingType
	}

	return nil
}

func telemetryHandler(pub Publisher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
			return
		}

		var telemetry Telemetry
		if err := json.NewDecoder(r.Body).Decode(&telemetry); err != nil {
			http.Error(w, "json invalido", http.StatusBadRequest)
			return
		}

		if err := validateTelemetryPayload(telemetry); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		body, err := json.Marshal(telemetry)
		if err != nil {
			http.Error(w, "erro ao serializar payload", http.StatusInternalServerError)
			return
		}

		err = pub.Publish(body)
		if err != nil {
			http.Error(w, "erro ao publicar mensagem", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("mensagem enfileirada com sucesso"))
	}
}
