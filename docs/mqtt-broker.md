# MQTT Broker

Field stations publish sensor readings directly to an embedded MQTT broker
built into the `cmd/broker` binary — there is no external Mosquitto (or
other) broker to stand up separately. The broker is implemented with
[`mochi-mqtt/server`](https://github.com/mochi-mqtt/server) and lives in
`internal/mqtt`.

## Running it

`make serve/dev` runs the broker alongside the web dashboard, both
hot-reloaded via `air` on `.go` file changes (the broker uses a separate
config, `.air.broker.toml`, so editing broker code doesn't rebuild the
dashboard and vice versa).

To run just the broker, without hot reload:

```
make broker/run       # go run ./cmd/broker
```

Requires `DATABASE_URL` (same as `cmd/web`/`cmd/seed`; see `.envrc`). By
default it listens on `:1883`; override with `MQTT_ADDR`:

```
MQTT_ADDR=:1884 make broker/run
```

There is currently no authentication — any client can connect and publish
(equivalent to the old Mosquitto config's `allow_anonymous true`). Don't
expose the broker port on an untrusted network without adding auth first.

## Logging

The broker logs structured JSON to stdout via `log/slog`, one object per
line. This covers both the broker's own protocol-level events (listener
startup, client connect/disconnect) and this app's ingestion events
(drops, successful inserts) — everything goes through the same logger, so
`grep`/`jq` over the process's stdout is enough to debug either layer.

Control verbosity with `LOG_LEVEL` (`debug`, `info`, `warn`, `error`;
default `info`; case-insensitive):

```
LOG_LEVEL=debug make broker/run
```

At `info` (the default) you'll see startup/shutdown and every dropped
reading. `debug` additionally logs every successful insert
(`"msg":"reading ingested"`) and the broker's own connection-level chatter
— useful when tracing why a specific station's readings aren't showing up,
noisy for normal operation. An unrecognized `LOG_LEVEL` value fails fast at
startup rather than silently falling back to a default.

Example lines:

```json
{"time":"...","level":"WARN","msg":"unknown station, dropping reading","station_id":"stn-typo","topic":"stn-typo/humidity"}
{"time":"...","level":"DEBUG","msg":"reading ingested","station_id":"stn-seed-001","type":"humidity","value":61.4,"recorded_at":"..."}
```

## Connecting

Any standard MQTT client library works — the broker (mochi-mqtt v2) is
compliant with MQTT v3.0.0, v3.1.1, and v5.

| Parameter        | Value                                                        |
|-------------------|---------------------------------------------------------------|
| Host              | wherever `cmd/broker` runs (`localhost` for local dev)        |
| Port              | `1883` (or whatever `MQTT_ADDR` is set to)                    |
| Transport         | plain TCP, no TLS                                              |
| Protocol version  | 3.1.1 or 5 both work; use whatever your client library defaults to |
| Auth              | none — any username/password or none at all is accepted        |
| Client ID         | any non-empty string; doesn't need to be pre-registered        |
| Clean session     | either works — the broker doesn't use subscriptions for ingestion |
| QoS               | 0 or 1 both work; the broker doesn't publish anything back, so there's nothing to subscribe to |

Connect, then publish to `<station-id>/<reading-type>` (see contracts
below) for each reading. There's no handshake beyond the standard MQTT
CONNECT — once connected, just publish.

## Topic contract

Publish to:

```
<station-id>/<reading-type>
```

`<station-id>` must already exist in the `stations` table (created via the
dashboard's "add station" form, or `make db/seed`) — readings for unknown
station IDs are logged and silently dropped, not queued or retried.

`<reading-type>` must be one of:

| Topic segment    | Stored as (`reading_type` enum) | Value                                    |
|------------------|---------------------------------|-------------------------------------------|
| `temperature`    | `temperature`                   | degrees Fahrenheit (dashboard renders `°F`, see `cmd/web/handler.go`) |
| `humidity`       | `humidity`                      | 0-100 percent                              |
| `soil-moisture`  | `soil_moisture`                 | 0-100 percent                              |
| `wind-direction` | `wind_direction`                | 0-360 degrees                              |

These ranges are conventions expected by the dashboard, not enforced by the
broker — `value` is stored as-is with no bounds checking, so out-of-range or
wrong-unit values won't be rejected, they'll just render oddly.

Any other segment, or a topic that isn't exactly two `/`-separated parts,
is logged and dropped.

## Payload contract

JSON body:

```json
{"value": 42.7, "collected_at": "2026-07-15T14:32:00Z"}
```

| Field          | Required | Notes                                                                 |
|-----------------|----------|------------------------------------------------------------------------|
| `value`         | yes      | `float64`, stored as-is in `readings.value`.                          |
| `collected_at`  | no       | RFC3339 timestamp. Omit it to have the server stamp receipt time. If present, it must parse as RFC3339 — an unparseable value drops the whole reading rather than falling back silently. |

## Client examples

### CLI (quick test)

```
mosquitto_pub -h localhost -p 1883 \
  -t stn-seed-001/humidity \
  -m '{"value":55.2,"collected_at":"2026-07-15T10:00:00Z"}'

# omit collected_at to use server receipt time
mosquitto_pub -h localhost -p 1883 \
  -t stn-seed-001/soil-moisture \
  -m '{"value":40.1}'
```

### Python (`paho-mqtt`)

```python
import json
from datetime import datetime, timezone
import paho.mqtt.client as mqtt

client = mqtt.Client()
client.connect("localhost", 1883)

payload = {
    "value": 55.2,
    "collected_at": datetime.now(timezone.utc).isoformat(),
}
client.publish("stn-seed-001/humidity", json.dumps(payload), qos=1)

client.disconnect()
```

### Node.js (`mqtt`)

```javascript
const mqtt = require("mqtt");
const client = mqtt.connect("mqtt://localhost:1883");

client.on("connect", () => {
  const payload = JSON.stringify({
    value: 55.2,
    collected_at: new Date().toISOString(),
  });
  client.publish("stn-seed-001/humidity", payload, { qos: 1 }, () => {
    client.end();
  });
});
```

### Arduino / ESP32 (`PubSubClient`)

```cpp
#include <WiFi.h>
#include <PubSubClient.h>

WiFiClient espClient;
PubSubClient client(espClient);

void publishReading(const char *stationID, const char *readingType, float value) {
  char topic[64];
  snprintf(topic, sizeof(topic), "%s/%s", stationID, readingType);

  // collected_at is optional — omit it to let the server stamp receipt time,
  // which is the simplest option if the device has no reliable RTC/NTP sync.
  char payload[64];
  snprintf(payload, sizeof(payload), "{\"value\":%.2f}", value);

  client.publish(topic, payload);
}

void setup() {
  client.setServer("192.168.1.50", 1883); // broker host/IP
  while (!client.connected()) {
    client.connect("field-station-01"); // any client ID
  }
}

void loop() {
  publishReading("stn-seed-001", "soil-moisture", 40.1);
  delay(60000);
}
```

## Drop semantics

The broker never returns an ingestion error to the publishing client — MQTT
has no request/response error channel for `PUBLISH`. Instead, every failure
mode is logged server-side and the reading is dropped:

- Unrecognized topic shape or reading-type segment
- Malformed JSON payload
- `collected_at` present but not valid RFC3339
- Unknown `station-id` (not yet created via the dashboard)
- Database insert failure

Check the broker's logs (stdout) if readings don't appear on the dashboard.

## Testing

`internal/mqtt/ingest_test.go` unit-tests the topic/payload parsing and
drop/insert logic directly against a mock `database.Querier`, without
needing a live broker connection — see that file for the exact cases
covered. Run with:

```
go test ./internal/mqtt/...
```
