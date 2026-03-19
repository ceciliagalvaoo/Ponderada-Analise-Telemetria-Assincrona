package main

import "time"

type DBInserter interface {
	InsertTelemetry(deviceID string, timestamp time.Time, sensorType, readingType string, value float64) error
}
