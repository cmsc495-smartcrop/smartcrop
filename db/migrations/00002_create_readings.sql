-- +goose Up
CREATE TABLE readings (
    id          BIGSERIAL   PRIMARY KEY,
    station_id  TEXT        NOT NULL REFERENCES stations(id),
    soil_moisture FLOAT,
    temperature FLOAT,
    humidity    FLOAT,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX readings_station_time ON readings (station_id, recorded_at DESC);

-- +goose Down
DROP TABLE readings;
