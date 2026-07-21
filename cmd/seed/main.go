package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Change this to seed more or fewer days of historical data
const SEED_DAYS = 90

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func vary(base, r float64) float64 {
	return base + rand.Float64()*r*2 - r
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

	type stationConfig struct {
		id       string
		name     string
		lat      float64
		lng      float64
		tempBase float64
		humBase  float64
		soilBase float64
		windBase float64
	}

	stations := []stationConfig{
		{"stn-seed-001", "North Field A",        38.5767, -93.2650, 76, 62, 44, 215},
		{"stn-seed-002", "South Greenhouse",     37.0902, -95.7129, 83, 72, 58, 46},
		{"stn-seed-003", "East Orchard Row 3",   35.4676, -97.5164, 90, 39, 30, 130},
		{"stn-seed-004", "West Irrigation Zone", 36.1540, -94.1323, 67, 56, 67, 312},
	}

	for _, s := range stations {
		if _, err := q.UpsertStation(ctx, database.UpsertStationParams{
			ID: s.id, Name: s.name, Latitude: s.lat, Longitude: s.lng,
		}); err != nil {
			return fmt.Errorf("upsert station %s: %w", s.id, err)
		}
		fmt.Printf("upserted station %s (%s)\n", s.id, s.name)

		count := 0

		for day := SEED_DAYS; day >= 1; day-- {
			for _, hour := range []int{2, 8, 14, 20} {
				ago := time.Duration(day*24*60+hour*60) * time.Minute
				for _, r := range []struct {
					t database.ReadingType
					v float64
				}{
					{database.ReadingTypeTemperature,   vary(s.tempBase, 3)},
					{database.ReadingTypeHumidity,      vary(s.humBase, 3)},
					{database.ReadingTypeSoilMoisture,  vary(s.soilBase, 2)},
					{database.ReadingTypeWindDirection,  vary(s.windBase, 15)},
				} {
					ts := pgtype.Timestamptz{Time: time.Now().Add(-ago), Valid: true}
					if _, err := q.CreateReading(ctx, database.CreateReadingParams{
						StationID: s.id, Type: r.t, Value: r.v, RecordedAt: ts,
					}); err != nil {
						return fmt.Errorf("create reading for %s: %w", s.id, err)
					}
					count++
				}
			}
		}

		// Recent readings for today
		for _, mins := range []int{32, 17, 2} {
			ago := time.Duration(mins) * time.Minute
			for _, r := range []struct {
				t database.ReadingType
				v float64
			}{
				{database.ReadingTypeTemperature,   vary(s.tempBase, 2)},
				{database.ReadingTypeHumidity,      vary(s.humBase, 2)},
				{database.ReadingTypeSoilMoisture,  vary(s.soilBase, 1)},
				{database.ReadingTypeWindDirection,  vary(s.windBase, 10)},
			} {
				ts := pgtype.Timestamptz{Time: time.Now().Add(-ago), Valid: true}
				if _, err := q.CreateReading(ctx, database.CreateReadingParams{
					StationID: s.id, Type: r.t, Value: r.v, RecordedAt: ts,
				}); err != nil {
					return fmt.Errorf("create reading for %s: %w", s.id, err)
				}
				count++
			}
		}

		fmt.Printf("inserted %d readings for %s\n", count, s.id)
	}

	fmt.Println("seed complete")
	return nil
}