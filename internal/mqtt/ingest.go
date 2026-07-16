package mqtt

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/packets"
)

// ReadingPayload is the wire format published to <station-id>/<reading-type>.
type ReadingPayload struct {
	Value       float64 `json:"value"`
	CollectedAt string  `json:"collected_at,omitempty"`
}

// topicReadingTypes maps the hyphenated topic segment to the DB enum.
var topicReadingTypes = map[string]database.ReadingType{
	"temperature":    database.ReadingTypeTemperature,
	"humidity":       database.ReadingTypeHumidity,
	"soil-moisture":  database.ReadingTypeSoilMoisture,
	"wind-direction": database.ReadingTypeWindDirection,
}

// parseTopic splits "<station-id>/<reading-type>" and maps the type segment
// to its database.ReadingType. ok is false if the topic isn't exactly two
// segments or the reading-type segment isn't recognized.
func parseTopic(topic string) (stationID string, typ database.ReadingType, ok bool) {
	parts := strings.SplitN(topic, "/", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", false
	}
	t, known := topicReadingTypes[parts[1]]
	if !known {
		return "", "", false
	}
	return parts[0], t, true
}

// ingestReading parses topic+payload and, if valid and the station is known,
// writes a reading via q. Malformed input or an unknown station/topic is
// logged and dropped rather than returned as an error, since there is no
// MQTT client to report failures back to.
func ingestReading(ctx context.Context, q database.Querier, logger *slog.Logger, topic string, payload []byte) {
	stationID, typ, ok := parseTopic(topic)
	if !ok {
		logger.Warn("unrecognized topic, dropping", "topic", topic)
		return
	}

	var p ReadingPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		logger.Warn("malformed payload, dropping", "topic", topic, "error", err)
		return
	}

	recordedAt := time.Now()
	if p.CollectedAt != "" {
		t, err := time.Parse(time.RFC3339, p.CollectedAt)
		if err != nil {
			logger.Warn("unparseable collected_at, dropping",
				"topic", topic, "collected_at", p.CollectedAt, "error", err)
			return
		}
		recordedAt = t
	}

	if _, err := q.GetStation(ctx, stationID); err != nil {
		logger.Warn("unknown station, dropping reading", "station_id", stationID, "topic", topic)
		return
	}

	if _, err := q.CreateReading(ctx, database.CreateReadingParams{
		StationID:  stationID,
		Type:       typ,
		Value:      p.Value,
		RecordedAt: pgtype.Timestamptz{Time: recordedAt, Valid: true},
	}); err != nil {
		logger.Error("insert reading failed",
			"station_id", stationID, "type", typ, "error", err)
		return
	}

	logger.Debug("reading ingested",
		"station_id", stationID, "type", typ, "value", p.Value, "recorded_at", recordedAt)
}

// ingestHook adapts the broker's OnPublish event to ingestReading.
type ingestHook struct {
	mochi.HookBase
	queries database.Querier
	logger  *slog.Logger
}

func (h *ingestHook) ID() string {
	return "smartcrop-ingest"
}

func (h *ingestHook) Provides(b byte) bool {
	return b == mochi.OnPublish
}

func (h *ingestHook) OnPublish(_ *mochi.Client, pk packets.Packet) (packets.Packet, error) {
	ingestReading(context.Background(), h.queries, h.logger, pk.TopicName, pk.Payload)
	return pk, nil
}
