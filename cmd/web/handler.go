package main

import "github.com/labstack/echo/v5"

func registerHandlers(e *echo.Echo) {
	h := &Handler{}
	e.GET("/", h.Home)
	e.GET("/station/:id", h.StationView)
}

type Handler struct{}

func (h *Handler) Home(c *echo.Context) error {
	return c.Render(200, "pages/home.gohtml", nil)
}

type StationView struct {
	StationID       string  `json:"station_id"`
	Temperature     float64 `json:"temperature"`
	TemperatureUnit string  `json:"temperature_unit"`
	Humidity        float64 `json:"humidity"`
}

func (h *Handler) StationView(c *echo.Context) error {
	stationID := c.Param("id")
	return c.Render(200, "pages/station.gohtml", StationView{StationID: stationID, Temperature: 70.0, TemperatureUnit: "F", Humidity: 50.0})
}
