package main

import "errors"

var (
	ErrInvalidDeviceID    = errors.New("device_id obrigatorio")
	ErrInvalidTimestamp   = errors.New("timestamp invalido: use RFC3339")
	ErrInvalidSensorType  = errors.New("sensor_type obrigatorio")
	ErrInvalidReadingType = errors.New("reading_type invalido: use analog ou discrete")
)
