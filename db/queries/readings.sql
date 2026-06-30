-- name: CreateReading :one
INSERT INTO readings (station_id, temperature, humidity, recorded_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLatestReading :one
SELECT * FROM readings
WHERE station_id = $1
ORDER BY recorded_at DESC
LIMIT 1;

-- name: ListReadingsByStation :many
SELECT * FROM readings
WHERE station_id = $1
ORDER BY recorded_at DESC
LIMIT $2;
