package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/labstack/echo/v5"
)

func registerHandlers(e *echo.Echo, queries *database.Queries) {
	h := &Handler{queries: queries}
	e.GET("/", h.HomeView)
	e.GET("/station/:id", h.StationView)
	e.POST("/stations", h.CreateStation)
}

func generateStationID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return "stn-" + hex.EncodeToString(b)
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

func formatLocation(lat, lng float64) string {
	alat, alng := lat, lng
	latDir, lngDir := "N", "W"
	if alat < 0 {
		alat, latDir = -alat, "S"
	}
	if alng >= 0 {
		lngDir = "E"
	} else {
		alng = -alng
	}
	return fmt.Sprintf("%.4f° %s, %.4f° %s", alat, latDir, alng, lngDir)
}

func stationListItem(s database.Station) StationListItem {
	return StationListItem{
		ID:       s.ID,
		Name:     s.Name,
		Location: formatLocation(s.Latitude, s.Longitude),
		Lat:      s.Latitude,
		Lng:      s.Longitude,
	}
}

func (h *Handler) HomeView(c *echo.Context) error {
	view := c.QueryParam("view")
	if view != "map" && view != "list" {
		view = "list"
	}

	dbStations, err := h.queries.ListStations(c.Request().Context())
	if err != nil {
		return err
	}
	stations := make([]StationListItem, len(dbStations))
	for i, s := range dbStations {
		stations[i] = stationListItem(s)
	}

	data := HomeData{View: view, Stations: stations}

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

type CreateStationRequest struct {
	Name      string  `form:"name"`
	Latitude  float64 `form:"latitude"`
	Longitude float64 `form:"longitude"`
}

func (h *Handler) CreateStation(c *echo.Context) error {
	var req CreateStationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	station, err := h.queries.CreateStation(c.Request().Context(), database.CreateStationParams{
		ID:        generateStationID(),
		Name:      req.Name,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
	if err != nil {
		return err
	}

	c.Response().Header().Set("HX-Redirect", "/station/"+station.ID)
	return c.NoContent(http.StatusOK)
}

type ReadingValue struct {
	Value      float64
	RecordedAt string
}

type StationReadings struct {
	Temperature   *ReadingValue
	Humidity      *ReadingValue
	SoilMoisture  *ReadingValue
	WindDirection *ReadingValue
}

type StationViewData struct {
	StationID string
	Name      string
	Lat       float64
	Lng       float64
	Location  string
	Readings  StationReadings
}

func (h *Handler) StationView(c *echo.Context) error {
	ctx := c.Request().Context()
	station, err := h.queries.GetStation(ctx, c.Param("id"))
	if err != nil {
		return err
	}

	latestReadings, err := h.queries.GetLatestReadings(ctx, station.ID)
	if err != nil {
		return err
	}

	var readings StationReadings
	for _, r := range latestReadings {
		rv := &ReadingValue{
			Value:      r.Value,
			RecordedAt: r.RecordedAt.Time.Format("Jan 2, 15:04"),
		}
		switch r.Type {
		case database.ReadingTypeTemperature:
			readings.Temperature = rv
		case database.ReadingTypeHumidity:
			readings.Humidity = rv
		case database.ReadingTypeSoilMoisture:
			readings.SoilMoisture = rv
		case database.ReadingTypeWindDirection:
			readings.WindDirection = rv
		}
	}

	return c.Render(200, "pages/station.gohtml", StationViewData{
		StationID: station.ID,
		Name:      station.Name,
		Lat:       station.Latitude,
		Lng:       station.Longitude,
		Location:  formatLocation(station.Latitude, station.Longitude),
		Readings:  readings,
	})
}
