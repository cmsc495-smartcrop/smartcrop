package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/cmsc495-smartcrop/smartcrop/internal/weather"
	"github.com/cmsc495-smartcrop/smartcrop/ui"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Start() error {
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

	queries := database.New(pool)

	forecaster, err := weather.NewClient()
	if err != nil {
		return fmt.Errorf("weather client: %w", err)
	}

	e := echo.New()

	tmpl, err := NewTemplateCache()
	if err != nil {
		return fmt.Errorf("template cache: %w", err)
	}

	e.Renderer = tmpl

	e.StaticFS("/static", ui.StaticFS())

	e.Use(middleware.Gzip())
	e.Use(middleware.RequestLogger())

	registerHandlers(e, queries, forecaster)

	return e.Start(":8080")
}
