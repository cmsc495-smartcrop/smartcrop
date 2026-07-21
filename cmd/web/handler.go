package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/labstack/echo/v5"
)

func registerHandlers(e *echo.Echo, queries database.Querier) {
	h := &Handler{queries: queries}
	e.GET("/", h.HomeView)
	e.GET("/station/:id", h.StationView)
	e.GET("/station/:id/readings", h.StationReadingsChart)
	e.POST("/stations", h.CreateStation)
}

func generateStationID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return "stn-" + hex.EncodeToString(b)
}

type Handler struct {
	queries database.Querier
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

type ReadingsChartData struct {
	Type       string
	Label      string
	Unit       string
	LabelsJSON template.JS
	ValuesJSON template.JS
}

func parseDate(s string) (time.Time, bool) {
	if s == "" {
		return time.Time{}, false
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

func (h *Handler) StationReadingsChart(c *echo.Context) error {
	ctx := c.Request().Context()
	stationID := c.Param("id")

	rawType := c.QueryParam("type")
	var dbType database.ReadingType
	var label, unit string
	switch rawType {
	case "humidity":
		dbType = database.ReadingTypeHumidity
		label, unit = "Humidity", "%"
	case "soil_moisture":
		dbType = database.ReadingTypeSoilMoisture
		label, unit = "Soil Moisture", "%"
	case "wind_direction":
		dbType = database.ReadingTypeWindDirection
		label, unit = "Wind Direction", "°"
	default:
		rawType = "temperature"
		dbType = database.ReadingTypeTemperature
		label, unit = "Temperature", "°F"
	}

	startDate, hasStart := parseDate(c.QueryParam("start"))
	endDate, hasEnd := parseDate(c.QueryParam("end"))

	var readings []database.Reading
	var err error

	if hasStart || hasEnd {

		all, fetchErr := h.queries.ListReadingsByStationAndType(ctx, database.ListReadingsByStationAndTypeParams{
			StationID: stationID,
			Type:      dbType,
			Limit:     500,
		})
		if fetchErr != nil {
			return fetchErr
		}
		for _, r := range all {
			t := r.RecordedAt.Time
			if hasStart && t.Before(startDate) {
				continue
			}
			if hasEnd && t.After(endDate.Add(24*time.Hour)) {
				continue
			}
			readings = append(readings, r)
		}
	} else {
		readings, err = h.queries.ListReadingsByStationAndType(ctx, database.ListReadingsByStationAndTypeParams{
			StationID: stationID,
			Type:      dbType,
			Limit:     500,
		})
		if err != nil {
			return err
		}
	}

	// Reverse to chronological order (oldest first)
	for i, j := 0, len(readings)-1; i < j; i, j = i+1, j-1 {
		readings[i], readings[j] = readings[j], readings[i]
	}

	labels := make([]string, len(readings))
	values := make([]float64, len(readings))
	for i, r := range readings {
		labels[i] = r.RecordedAt.Time.Format("Jan 2, 15:04")
		values[i] = r.Value
	}

	labelsJSON, _ := json.Marshal(labels)
	valuesJSON, _ := json.Marshal(values)

	return c.Render(200, "partials/readings_chart.gohtml", ReadingsChartData{
		Type:       rawType,
		Label:      label,
		Unit:       unit,
		LabelsJSON: template.JS(labelsJSON),
		ValuesJSON: template.JS(valuesJSON),
	})
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