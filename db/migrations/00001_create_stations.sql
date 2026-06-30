-- +goose Up
CREATE TABLE stations (
    id         TEXT        PRIMARY KEY,
    name       TEXT        NOT NULL,
    latitude double precision NOT NULL,
    longitude double precision NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE stations;
