package main

import (
	"encoding/json"
	"time"
)

func processMessage(repo DBInserter, body []byte) error {
	var telemetry Telemetry
	if err := json.Unmarshal(body, &telemetry); err != nil {
		return err
	}

	t, err := time.Parse(time.RFC3339, telemetry.Timestamp)
	if err != nil {
		return err
	}

	return repo.InsertTelemetry(
		telemetry.DeviceID,
		t,
		telemetry.SensorType,
		telemetry.ReadingType,
		telemetry.Value,
	)
}
