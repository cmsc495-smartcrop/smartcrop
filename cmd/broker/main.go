package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/cmsc495-smartcrop/smartcrop/internal/mqtt"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	logLevel := new(slog.LevelVar)
	if lvl := os.Getenv("LOG_LEVEL"); lvl != "" {
		if err := logLevel.UnmarshalText([]byte(lvl)); err != nil {
			return fmt.Errorf("invalid LOG_LEVEL %q: %w", lvl, err)
		}
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

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

	addr := os.Getenv("MQTT_ADDR")
	if addr == "" {
		addr = ":1883"
	}

	queries := database.New(pool)
	broker, err := mqtt.NewBroker(addr, queries, logger)
	if err != nil {
		return fmt.Errorf("mqtt broker: %w", err)
	}

	if err := broker.Serve(); err != nil {
		return fmt.Errorf("mqtt broker: %w", err)
	}
	logger.Info("mqtt broker listening", "addr", addr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	logger.Info("shutting down mqtt broker")
	return broker.Close()
}
