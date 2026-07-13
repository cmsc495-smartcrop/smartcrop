package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx/v5/pgtype"
)

type SensorPayload struct {
	StationID    string  `json:"station_id"`
	Temperature  float64 `json:"temperature"`
	Humidity     float64 `json:"humidity"`
	SoilMoisture float64 `json:"soil_moisture"`
}

type Subscriber struct {
	client  paho.Client
	queries *database.Queries
}

func NewSubscriber(brokerURL string, queries *database.Queries) (*Subscriber, error) {
	opts := paho.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID("smartcrop-subscriber").
		SetAutoReconnect(true).
		SetConnectRetry(true)

	s := &Subscriber{queries: queries}

	opts.SetOnConnectHandler(func(c paho.Client) {
		log.Println("mqtt: connected to broker")
		c.Subscribe("stations/+/readings", 1, s.handleReading)
	})

	client := paho.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	s.client = client
	return s, nil
}

func (s *Subscriber) handleReading(_ paho.Client, msg paho.Message) {
	var p SensorPayload
	if err := json.Unmarshal(msg.Payload(), &p); err != nil {
		log.Printf("mqtt: malformed payload on %s: %v", msg.Topic(), err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := s.queries.GetStation(ctx, p.StationID); err != nil {
		log.Printf("mqtt: unknown station %q, dropping reading", p.StationID)
		return
	}

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	readings := []struct {
		t database.ReadingType
		v float64
	}{
		{database.ReadingTypeTemperature, p.Temperature},
		{database.ReadingTypeHumidity, p.Humidity},
		{database.ReadingTypeSoilMoisture, p.SoilMoisture},
	}
	for _, r := range readings {
		if _, err := s.queries.CreateReading(ctx, database.CreateReadingParams{
			StationID:  p.StationID,
			Type:       r.t,
			Value:      r.v,
			RecordedAt: now,
		}); err != nil {
			log.Printf("mqtt: insert %s reading for station %q: %v", r.t, p.StationID, err)
		}
	}
}

func (s *Subscriber) Close() {
	s.client.Disconnect(250)
}
