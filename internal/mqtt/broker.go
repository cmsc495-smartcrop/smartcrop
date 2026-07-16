// Package mqtt implements an embedded MQTT broker that field stations
// publish sensor readings to directly, ingesting them into the database.
package mqtt

import (
	"fmt"
	"log/slog"

	"github.com/cmsc495-smartcrop/smartcrop/internal/database"
	mochi "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
)

// Broker is an embedded MQTT broker that ingests sensor readings published
// to <station-id>/<reading-type> topics into the database.
type Broker struct {
	server *mochi.Server
}

// NewBroker constructs a broker listening on addr (e.g. ":1883"). It does
// not start accepting connections until Serve is called. logger receives
// both the broker's own protocol-level events (client connects, listener
// lifecycle, etc.) and this package's ingestion logs, so operators get one
// consistent, leveled log stream.
func NewBroker(addr string, queries database.Querier, logger *slog.Logger) (*Broker, error) {
	server := mochi.New(&mochi.Options{Logger: logger})

	// No authentication story exists yet; allow any client to connect and
	// publish, matching the previous Mosquitto config's allow_anonymous.
	if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
		return nil, fmt.Errorf("add allow-all auth hook: %w", err)
	}

	if err := server.AddHook(&ingestHook{queries: queries, logger: logger}, nil); err != nil {
		return nil, fmt.Errorf("add ingest hook: %w", err)
	}

	tcp := listeners.NewTCP(listeners.Config{ID: "tcp", Address: addr})
	if err := server.AddListener(tcp); err != nil {
		return nil, fmt.Errorf("add tcp listener on %s: %w", addr, err)
	}

	return &Broker{server: server}, nil
}

// Serve starts the broker's listeners in background goroutines and returns
// once they're up (or immediately with an error if setup failed). Call
// Close to shut the broker down.
func (b *Broker) Serve() error {
	return b.server.Serve()
}

// Close gracefully shuts down the broker.
func (b *Broker) Close() error {
	return b.server.Close()
}
