package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func connectDB() *sql.DB {
	connStr := "host=db port=5432 user=postgres password=postgres dbname=telemetrydb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("erro ao conectar no banco:", err)
	}

	for attempt := 1; attempt <= 120; attempt++ {
		err = db.Ping()
		if err == nil {
			return db
		}

		log.Printf("tentativa %d/120: erro ao pingar banco: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("erro ao pingar banco apos tentativas:", err)
	}

	return db
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (p *PostgresRepository) InsertTelemetry(deviceID string, timestamp time.Time, sensorType, readingType string, value float64) error {
	_, err := p.db.Exec(`
		INSERT INTO telemetry (device_id, timestamp, sensor_type, reading_type, value)
		VALUES ($1, $2, $3, $4, $5)
	`,
		deviceID,
		timestamp,
		sensorType,
		readingType,
		value,
	)
	return err
}
