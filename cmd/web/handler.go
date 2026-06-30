package main

import (
	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/labstack/echo/v5"
)

func registerHandlers(e *echo.Echo) {
	h := &Handler{}
	e.GET("/", h.HomeView)
	e.GET("/station/:id", h.StationView)
}

type Handler struct {
	queries *database.Queries
}

type StationListItem struct {
	ID       string
	Name     string
	Location string
	Lat      float64
	Lng      float64
}

type HomeData struct {
	View     string
	Stations []StationListItem
}

var sampleStations = []StationListItem{
	{ID: "stn-001", Name: "North Field A", Location: "38.5767° N, 93.2650° W", Lat: 38.5767, Lng: -93.2650},
	{ID: "stn-002", Name: "South Greenhouse", Location: "37.0902° N, 95.7129° W", Lat: 37.0902, Lng: -95.7129},
	{ID: "stn-003", Name: "East Orchard Row 3", Location: "35.4676° N, 97.5164° W", Lat: 35.4676, Lng: -97.5164},
	{ID: "stn-004", Name: "West Irrigation Zone", Location: "36.1540° N, 94.1323° W", Lat: 36.1540, Lng: -94.1323},
}

func (h *Handler) HomeView(c *echo.Context) error {
	view := c.QueryParam("view")
	if view != "map" && view != "list" {
		view = "list"
	}

	data := HomeData{View: view, Stations: sampleStations}

	if c.Request().Header.Get("HX-Request") == "true" {
		switch view {
		case "map":
			return c.Render(200, "partials/station_map.gohtml", data)
		default:
			return c.Render(200, "partials/station_list.gohtml", data)
		}
	}

	return c.Render(200, "pages/home.gohtml", data)
}

type StationViewData struct {
	StationID       string  `json:"station_id"`
	Temperature     float64 `json:"temperature"`
	TemperatureUnit string  `json:"temperature_unit"`
	Humidity        float64 `json:"humidity"`
}

func (h *Handler) StationView(c *echo.Context) error {
	stationID := c.Param("id")
	return c.Render(200, "pages/station.gohtml", StationViewData{StationID: stationID, Temperature: 70.0, TemperatureUnit: "F", Humidity: 50.0})
}
