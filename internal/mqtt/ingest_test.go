package mqtt

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
)

// testLogger discards output; tests assert on createReadingCalls, not logs.
var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

// mockQuerier is a controllable in-memory implementation of
// database.Querier, following the pattern used in cmd/web/handler_test.go.
type mockQuerier struct {
	station       database.Station
	getStationErr error

	createReadingCalls []database.CreateReadingParams
	createReadingErr   error
}

func (m *mockQuerier) GetStation(_ context.Context, _ string) (database.Station, error) {
	return m.station, m.getStationErr
}

func (m *mockQuerier) CreateReading(_ context.Context, arg database.CreateReadingParams) (database.Reading, error) {
	m.createReadingCalls = append(m.createReadingCalls, arg)
	return database.Reading{}, m.createReadingErr
}

func (m *mockQuerier) CreateStation(_ context.Context, _ database.CreateStationParams) (database.Station, error) {
	return database.Station{}, nil
}
func (m *mockQuerier) GetLatestReadingByType(_ context.Context, _ database.GetLatestReadingByTypeParams) (database.Reading, error) {
	return database.Reading{}, nil
}
func (m *mockQuerier) GetLatestReadings(_ context.Context, _ string) ([]database.Reading, error) {
	return nil, nil
}
func (m *mockQuerier) ListReadingsByStation(_ context.Context, _ database.ListReadingsByStationParams) ([]database.Reading, error) {
	return nil, nil
}
func (m *mockQuerier) ListReadingsByStationAndType(_ context.Context, _ database.ListReadingsByStationAndTypeParams) ([]database.Reading, error) {
	return nil, nil
}
func (m *mockQuerier) ListReadingsByStationAndTypeAndDateRange(_ context.Context, _ database.ListReadingsByStationAndTypeAndDateRangeParams) ([]database.Reading, error) {
	return nil, nil
}
func (m *mockQuerier) ListStations(_ context.Context) ([]database.Station, error) {
	return nil, nil
}
func (m *mockQuerier) UpsertStation(_ context.Context, _ database.UpsertStationParams) (database.Station, error) {
	return database.Station{}, nil
}

var _ database.Querier = (*mockQuerier)(nil)

func TestParseTopic(t *testing.T) {
	tests := []struct {
		topic    string
		wantOK   bool
		wantID   string
		wantType database.ReadingType
	}{
		{"stn-1/temperature", true, "stn-1", database.ReadingTypeTemperature},
		{"stn-1/humidity", true, "stn-1", database.ReadingTypeHumidity},
		{"stn-1/soil-moisture", true, "stn-1", database.ReadingTypeSoilMoisture},
		{"stn-1/wind-direction", true, "stn-1", database.ReadingTypeWindDirection},
		{"stn-1/pressure", false, "", ""},
		{"stn-1/soil_moisture", false, "", ""}, // underscore form not accepted on the wire
		{"stn-1", false, "", ""},
		{"stn-1/temperature/extra", false, "", ""},
		{"/temperature", false, "", ""},
		{"", false, "", ""},
	}
	for _, tc := range tests {
		gotID, gotType, gotOK := parseTopic(tc.topic)
		if gotOK != tc.wantOK || gotID != tc.wantID || gotType != tc.wantType {
			t.Errorf("parseTopic(%q) = (%q, %q, %v), want (%q, %q, %v)",
				tc.topic, gotID, gotType, gotOK, tc.wantID, tc.wantType, tc.wantOK)
		}
	}
}

func TestIngestReading_KnownTopicKnownStation_Inserts(t *testing.T) {
	q := &mockQuerier{station: database.Station{ID: "stn-1"}}

	ingestReading(context.Background(), q, testLogger, "stn-1/humidity", []byte(`{"value":55.2,"collected_at":"2026-07-15T10:00:00Z"}`))

	if len(q.createReadingCalls) != 1 {
		t.Fatalf("createReadingCalls = %d, want 1", len(q.createReadingCalls))
	}
	call := q.createReadingCalls[0]
	if call.StationID != "stn-1" || call.Type != database.ReadingTypeHumidity || call.Value != 55.2 {
		t.Errorf("unexpected call: %+v", call)
	}
	wantTime := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	if !call.RecordedAt.Valid || !call.RecordedAt.Time.Equal(wantTime) {
		t.Errorf("RecordedAt = %v, want %v", call.RecordedAt, wantTime)
	}
}

func TestIngestReading_MissingCollectedAt_DefaultsToNow(t *testing.T) {
	q := &mockQuerier{station: database.Station{ID: "stn-1"}}
	before := time.Now()

	ingestReading(context.Background(), q, testLogger, "stn-1/soil-moisture", []byte(`{"value":40.1}`))

	after := time.Now()
	if len(q.createReadingCalls) != 1 {
		t.Fatalf("createReadingCalls = %d, want 1", len(q.createReadingCalls))
	}
	got := q.createReadingCalls[0].RecordedAt.Time
	if got.Before(before) || got.After(after) {
		t.Errorf("RecordedAt = %v, want between %v and %v", got, before, after)
	}
}

func TestIngestReading_UnparseableCollectedAt_Drops(t *testing.T) {
	q := &mockQuerier{station: database.Station{ID: "stn-1"}}

	ingestReading(context.Background(), q, testLogger, "stn-1/temperature", []byte(`{"value":1,"collected_at":"not-a-time"}`))

	if len(q.createReadingCalls) != 0 {
		t.Errorf("createReadingCalls = %d, want 0", len(q.createReadingCalls))
	}
}

func TestIngestReading_UnknownStation_Drops(t *testing.T) {
	q := &mockQuerier{getStationErr: errors.New("station not found")}

	ingestReading(context.Background(), q, testLogger, "stn-does-not-exist/temperature", []byte(`{"value":1}`))

	if len(q.createReadingCalls) != 0 {
		t.Errorf("createReadingCalls = %d, want 0", len(q.createReadingCalls))
	}
}

func TestIngestReading_UnknownTopicSegment_Drops(t *testing.T) {
	q := &mockQuerier{station: database.Station{ID: "stn-1"}}

	ingestReading(context.Background(), q, testLogger, "stn-1/pressure", []byte(`{"value":1}`))

	if len(q.createReadingCalls) != 0 {
		t.Errorf("createReadingCalls = %d, want 0", len(q.createReadingCalls))
	}
}

func TestIngestReading_MalformedTopicShape_Drops(t *testing.T) {
	q := &mockQuerier{station: database.Station{ID: "stn-1"}}

	ingestReading(context.Background(), q, testLogger, "stn-1", []byte(`{"value":1}`))

	if len(q.createReadingCalls) != 0 {
		t.Errorf("createReadingCalls = %d, want 0", len(q.createReadingCalls))
	}
}

func TestIngestReading_MalformedJSON_Drops(t *testing.T) {
	q := &mockQuerier{station: database.Station{ID: "stn-1"}}

	ingestReading(context.Background(), q, testLogger, "stn-1/temperature", []byte(`not json`))

	if len(q.createReadingCalls) != 0 {
		t.Errorf("createReadingCalls = %d, want 0", len(q.createReadingCalls))
	}
}
