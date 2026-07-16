package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v5"
)

// testRenderer captures the template name and data without writing a body.
type testRenderer struct {
	lastTemplate string
	lastData     any
}

func (r *testRenderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	r.lastTemplate = name
	r.lastData = data
	return nil
}

// mockQuerier is a controllable in-memory implementation of database.Querier.
type mockQuerier struct {
	stations          []database.Station
	listStationsErr   error
	station           database.Station
	getStationErr     error
	createdStation    database.Station
	createStationErr  error
	latestReadings    []database.Reading
	latestReadingsErr error
	typeReadings      []database.Reading
	typeReadingsErr   error
}

func (m *mockQuerier) ListStations(_ context.Context) ([]database.Station, error) {
	return m.stations, m.listStationsErr
}
func (m *mockQuerier) GetStation(_ context.Context, _ string) (database.Station, error) {
	return m.station, m.getStationErr
}
func (m *mockQuerier) CreateStation(_ context.Context, _ database.CreateStationParams) (database.Station, error) {
	return m.createdStation, m.createStationErr
}
func (m *mockQuerier) GetLatestReadings(_ context.Context, _ string) ([]database.Reading, error) {
	return m.latestReadings, m.latestReadingsErr
}
func (m *mockQuerier) ListReadingsByStationAndType(_ context.Context, _ database.ListReadingsByStationAndTypeParams) ([]database.Reading, error) {
	return m.typeReadings, m.typeReadingsErr
}
func (m *mockQuerier) GetLatestReadingByType(_ context.Context, _ database.GetLatestReadingByTypeParams) (database.Reading, error) {
	return database.Reading{}, nil
}
func (m *mockQuerier) CreateReading(_ context.Context, _ database.CreateReadingParams) (database.Reading, error) {
	return database.Reading{}, nil
}
func (m *mockQuerier) ListReadingsByStation(_ context.Context, _ database.ListReadingsByStationParams) ([]database.Reading, error) {
	return nil, nil
}
func (m *mockQuerier) UpsertStation(_ context.Context, _ database.UpsertStationParams) (database.Station, error) {
	return database.Station{}, nil
}

func pgts(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func newTestEcho(rend *testRenderer, q database.Querier) *echo.Echo {
	e := echo.New()
	e.Renderer = rend
	registerHandlers(e, q)
	return e
}

// --- formatLocation ---

func TestFormatLocation(t *testing.T) {
	tests := []struct {
		lat, lng float64
		want     string
	}{
		{38.5767, -93.2650, "38.5767° N, 93.2650° W"},
		{-33.8688, 151.2093, "33.8688° S, 151.2093° E"},
		{0, 0, "0.0000° N, 0.0000° E"},
		{-90, -180, "90.0000° S, 180.0000° W"},
	}
	for _, tc := range tests {
		got := formatLocation(tc.lat, tc.lng)
		if got != tc.want {
			t.Errorf("formatLocation(%v, %v) = %q, want %q", tc.lat, tc.lng, got, tc.want)
		}
	}
}

// --- HomeView ---

func TestHomeView_DefaultsList(t *testing.T) {
	rend := &testRenderer{}
	mock := &mockQuerier{
		stations: []database.Station{
			{ID: "stn-1", Name: "Field A", Latitude: 38.5767, Longitude: -93.2650},
		},
	}
	e := newTestEcho(rend, mock)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rend.lastTemplate != "pages/home.gohtml" {
		t.Errorf("template = %q, want pages/home.gohtml", rend.lastTemplate)
	}
	data := rend.lastData.(HomeData)
	if data.View != "list" {
		t.Errorf("view = %q, want list", data.View)
	}
	if len(data.Stations) != 1 {
		t.Errorf("stations = %d, want 1", len(data.Stations))
	}
}

func TestHomeView_MapView(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/?view=map", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(HomeData)
	if data.View != "map" {
		t.Errorf("view = %q, want map", data.View)
	}
	if rend.lastTemplate != "pages/home.gohtml" {
		t.Errorf("template = %q, want pages/home.gohtml", rend.lastTemplate)
	}
}

func TestHomeView_InvalidViewDefaultsList(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/?view=bogus", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(HomeData)
	if data.View != "list" {
		t.Errorf("view = %q, want list", data.View)
	}
}

func TestHomeView_HTMXListPartial(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/?view=list", nil)
	req.Header.Set("HX-Request", "true")
	e.ServeHTTP(httptest.NewRecorder(), req)

	if rend.lastTemplate != "partials/station_list.gohtml" {
		t.Errorf("template = %q, want partials/station_list.gohtml", rend.lastTemplate)
	}
}

func TestHomeView_HTMXMapPartial(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/?view=map", nil)
	req.Header.Set("HX-Request", "true")
	e.ServeHTTP(httptest.NewRecorder(), req)

	if rend.lastTemplate != "partials/station_map.gohtml" {
		t.Errorf("template = %q, want partials/station_map.gohtml", rend.lastTemplate)
	}
}

func TestHomeView_StationFormatting(t *testing.T) {
	rend := &testRenderer{}
	mock := &mockQuerier{
		stations: []database.Station{
			{ID: "stn-1", Name: "Field A", Latitude: 38.5767, Longitude: -93.2650},
		},
	}
	e := newTestEcho(rend, mock)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	item := rend.lastData.(HomeData).Stations[0]
	if item.ID != "stn-1" || item.Name != "Field A" {
		t.Errorf("station identity wrong: %+v", item)
	}
	if item.Lat != 38.5767 || item.Lng != -93.2650 {
		t.Errorf("lat/lng wrong: %v, %v", item.Lat, item.Lng)
	}
	want := "38.5767° N, 93.2650° W"
	if item.Location != want {
		t.Errorf("location = %q, want %q", item.Location, want)
	}
}

// --- StationView ---

func TestStationView_RendersStation(t *testing.T) {
	rend := &testRenderer{}
	mock := &mockQuerier{
		station: database.Station{ID: "stn-1", Name: "Field A", Latitude: 38.5767, Longitude: -93.2650},
	}
	e := newTestEcho(rend, mock)

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if rend.lastTemplate != "pages/station.gohtml" {
		t.Errorf("template = %q, want pages/station.gohtml", rend.lastTemplate)
	}
	data := rend.lastData.(StationViewData)
	if data.StationID != "stn-1" || data.Name != "Field A" {
		t.Errorf("station identity wrong: %+v", data)
	}
	if data.Location != "38.5767° N, 93.2650° W" {
		t.Errorf("location = %q", data.Location)
	}
}

func TestStationView_ReadingsMappedByType(t *testing.T) {
	now := time.Now()
	rend := &testRenderer{}
	mock := &mockQuerier{
		station: database.Station{ID: "stn-1"},
		latestReadings: []database.Reading{
			{Type: database.ReadingTypeTemperature, Value: 78.4, RecordedAt: pgts(now)},
			{Type: database.ReadingTypeHumidity, Value: 61.2, RecordedAt: pgts(now)},
			{Type: database.ReadingTypeSoilMoisture, Value: 43.7, RecordedAt: pgts(now)},
			{Type: database.ReadingTypeWindDirection, Value: 225, RecordedAt: pgts(now)},
		},
	}
	e := newTestEcho(rend, mock)

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	r := rend.lastData.(StationViewData).Readings
	if r.Temperature == nil || r.Temperature.Value != 78.4 {
		t.Error("temperature not mapped correctly")
	}
	if r.Humidity == nil || r.Humidity.Value != 61.2 {
		t.Error("humidity not mapped correctly")
	}
	if r.SoilMoisture == nil || r.SoilMoisture.Value != 43.7 {
		t.Error("soil moisture not mapped correctly")
	}
	if r.WindDirection == nil || r.WindDirection.Value != 225 {
		t.Error("wind direction not mapped correctly")
	}
}

func TestStationView_NoReadings(t *testing.T) {
	rend := &testRenderer{}
	mock := &mockQuerier{station: database.Station{ID: "stn-1"}}
	e := newTestEcho(rend, mock)

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	r := rend.lastData.(StationViewData).Readings
	if r.Temperature != nil || r.Humidity != nil || r.SoilMoisture != nil || r.WindDirection != nil {
		t.Error("expected all readings to be nil")
	}
}

// --- StationReadingsChart ---

func TestStationReadingsChart_DefaultsToTemperature(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1/readings", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(ReadingsChartData)
	if data.Type != "temperature" {
		t.Errorf("type = %q, want temperature", data.Type)
	}
	if data.Unit != "°F" {
		t.Errorf("unit = %q, want °F", data.Unit)
	}
}

func TestStationReadingsChart_HumidityType(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1/readings?type=humidity", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(ReadingsChartData)
	if data.Type != "humidity" || data.Unit != "%" || data.Label != "Humidity" {
		t.Errorf("got type=%q unit=%q label=%q", data.Type, data.Unit, data.Label)
	}
}

func TestStationReadingsChart_SoilMoisture(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1/readings?type=soil_moisture", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(ReadingsChartData)
	if data.Type != "soil_moisture" || data.Label != "Soil Moisture" {
		t.Errorf("got type=%q label=%q", data.Type, data.Label)
	}
}

func TestStationReadingsChart_WindDirection(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1/readings?type=wind_direction", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(ReadingsChartData)
	if data.Type != "wind_direction" || data.Unit != "°" {
		t.Errorf("got type=%q unit=%q", data.Type, data.Unit)
	}
}

func TestStationReadingsChart_UnknownTypeDefaultsToTemperature(t *testing.T) {
	rend := &testRenderer{}
	e := newTestEcho(rend, &mockQuerier{})

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1/readings?type=pressure", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(ReadingsChartData)
	if data.Type != "temperature" {
		t.Errorf("type = %q, want temperature", data.Type)
	}
}

func TestStationReadingsChart_ChronologicalOrder(t *testing.T) {
	older := time.Now().Add(-30 * time.Minute)
	newer := time.Now()

	rend := &testRenderer{}
	mock := &mockQuerier{
		// DB returns newest-first; handler should reverse to oldest-first.
		typeReadings: []database.Reading{
			{Type: database.ReadingTypeTemperature, Value: 80.0, RecordedAt: pgts(newer)},
			{Type: database.ReadingTypeTemperature, Value: 75.0, RecordedAt: pgts(older)},
		},
	}
	e := newTestEcho(rend, mock)

	req := httptest.NewRequest(http.MethodGet, "/station/stn-1/readings", nil)
	e.ServeHTTP(httptest.NewRecorder(), req)

	data := rend.lastData.(ReadingsChartData)
	// Oldest value (75) should appear before newest (80).
	if string(data.ValuesJSON) != "[75,80]" {
		t.Errorf("ValuesJSON = %s, want [75,80]", data.ValuesJSON)
	}
}

// --- CreateStation ---

func TestCreateStation_Success(t *testing.T) {
	rend := &testRenderer{}
	mock := &mockQuerier{
		createdStation: database.Station{ID: "stn-abc123"},
	}
	e := newTestEcho(rend, mock)

	form := url.Values{}
	form.Set("name", "Test Field")
	form.Set("latitude", "38.5767")
	form.Set("longitude", "-93.2650")
	req := httptest.NewRequest(http.MethodPost, "/stations", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("HX-Redirect"); got != "/station/stn-abc123" {
		t.Errorf("HX-Redirect = %q, want /station/stn-abc123", got)
	}
}
