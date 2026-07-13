package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return fmt.Errorf("database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		return fmt.Errorf("database ping: %w", err)
	}

	q := database.New(pool)
	ctx := context.Background()

	stations := []database.UpsertStationParams{
		{ID: "stn-seed-001", Name: "North Field A", Latitude: 38.5767, Longitude: -93.2650},
		{ID: "stn-seed-002", Name: "South Greenhouse", Latitude: 37.0902, Longitude: -95.7129},
		{ID: "stn-seed-003", Name: "East Orchard Row 3", Latitude: 35.4676, Longitude: -97.5164},
		{ID: "stn-seed-004", Name: "West Irrigation Zone", Latitude: 36.1540, Longitude: -94.1323},
	}

	type reading struct {
		typ   database.ReadingType
		value float64
		ago   time.Duration
	}

	stationReadings := map[string][]reading{
		"stn-seed-001": {
			{database.ReadingTypeTemperature, 78.4, 2 * time.Minute},
			{database.ReadingTypeTemperature, 77.9, 17 * time.Minute},
			{database.ReadingTypeTemperature, 76.1, 32 * time.Minute},
			{database.ReadingTypeHumidity, 61.2, 2 * time.Minute},
			{database.ReadingTypeHumidity, 62.5, 17 * time.Minute},
			{database.ReadingTypeHumidity, 64.0, 32 * time.Minute},
			{database.ReadingTypeSoilMoisture, 43.7, 2 * time.Minute},
			{database.ReadingTypeSoilMoisture, 44.1, 17 * time.Minute},
			{database.ReadingTypeSoilMoisture, 44.8, 32 * time.Minute},
			{database.ReadingTypeWindDirection, 225, 2 * time.Minute},
			{database.ReadingTypeWindDirection, 218, 17 * time.Minute},
			{database.ReadingTypeWindDirection, 230, 32 * time.Minute},
		},
		"stn-seed-002": {
			{database.ReadingTypeTemperature, 84.1, 3 * time.Minute},
			{database.ReadingTypeTemperature, 83.6, 18 * time.Minute},
			{database.ReadingTypeTemperature, 82.0, 33 * time.Minute},
			{database.ReadingTypeHumidity, 72.8, 3 * time.Minute},
			{database.ReadingTypeHumidity, 71.3, 18 * time.Minute},
			{database.ReadingTypeHumidity, 70.5, 33 * time.Minute},
			{database.ReadingTypeSoilMoisture, 58.2, 3 * time.Minute},
			{database.ReadingTypeSoilMoisture, 57.9, 18 * time.Minute},
			{database.ReadingTypeSoilMoisture, 57.4, 33 * time.Minute},
			{database.ReadingTypeWindDirection, 45, 3 * time.Minute},
			{database.ReadingTypeWindDirection, 52, 18 * time.Minute},
			{database.ReadingTypeWindDirection, 40, 33 * time.Minute},
		},
		"stn-seed-003": {
			{database.ReadingTypeTemperature, 91.3, 1 * time.Minute},
			{database.ReadingTypeTemperature, 90.7, 16 * time.Minute},
			{database.ReadingTypeHumidity, 38.4, 1 * time.Minute},
			{database.ReadingTypeHumidity, 39.1, 16 * time.Minute},
			{database.ReadingTypeSoilMoisture, 29.6, 1 * time.Minute},
			{database.ReadingTypeSoilMoisture, 30.2, 16 * time.Minute},
			{database.ReadingTypeWindDirection, 135, 1 * time.Minute},
			{database.ReadingTypeWindDirection, 128, 16 * time.Minute},
		},
		"stn-seed-004": {
			{database.ReadingTypeTemperature, 67.5, 5 * time.Minute},
			{database.ReadingTypeTemperature, 66.8, 20 * time.Minute},
			{database.ReadingTypeHumidity, 55.3, 5 * time.Minute},
			{database.ReadingTypeHumidity, 56.0, 20 * time.Minute},
			{database.ReadingTypeSoilMoisture, 66.9, 5 * time.Minute},
			{database.ReadingTypeSoilMoisture, 67.4, 20 * time.Minute},
			{database.ReadingTypeWindDirection, 315, 5 * time.Minute},
			{database.ReadingTypeWindDirection, 310, 20 * time.Minute},
		},
	}

	for _, s := range stations {
		if _, err := q.UpsertStation(ctx, s); err != nil {
			return fmt.Errorf("upsert station %s: %w", s.ID, err)
		}
		fmt.Printf("upserted station %s (%s)\n", s.ID, s.Name)

		for _, r := range stationReadings[s.ID] {
			ts := pgtype.Timestamptz{Time: time.Now().Add(-r.ago), Valid: true}
			if _, err := q.CreateReading(ctx, database.CreateReadingParams{
				StationID:  s.ID,
				Type:       r.typ,
				Value:      r.value,
				RecordedAt: ts,
			}); err != nil {
				return fmt.Errorf("create reading for %s: %w", s.ID, err)
			}
		}
		fmt.Printf("inserted %d readings for %s\n", len(stationReadings[s.ID]), s.ID)
	}

	fmt.Println("seed complete")
	return nil
}
