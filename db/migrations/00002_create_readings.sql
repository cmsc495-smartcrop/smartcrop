-- +goose Up
CREATE TYPE reading_type AS ENUM ('soil_moisture', 'temperature', 'humidity');

CREATE TABLE readings (
    id          BIGSERIAL    PRIMARY KEY,
    station_id  TEXT         NOT NULL REFERENCES stations(id),
    type        reading_type NOT NULL,
    value       FLOAT        NOT NULL,
    recorded_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX readings_station_time ON readings (station_id, recorded_at DESC);

-- +goose Down
DROP TABLE readings;
DROP TYPE reading_type;
