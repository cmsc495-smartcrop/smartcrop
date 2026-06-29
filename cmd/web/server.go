package main

import (
	"fmt"
	"smartcrop/ui"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func Start() error {
	e := echo.New()

	tmpl, err := NewTemplateCache()
	if err != nil {
		return fmt.Errorf("template cache: %w", err)
	}

	e.Renderer = tmpl

	e.StaticFS("/static", ui.StaticFS())

	e.Use(middleware.Gzip())
	e.Use(middleware.RequestLogger())

	registerHandlers(e)

	return e.Start(":8080")
}
