package main

import "github.com/labstack/echo/v5"

func registerHandlers(e *echo.Echo) {
	h := &Handler{}
	e.GET("/", h.Home)
}

type Handler struct{}

func (h *Handler) Home(c *echo.Context) error {
	return c.Render(200, "pages/home.gohtml", nil)
}
