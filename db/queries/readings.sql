-- name: CreateReading :one
INSERT INTO readings (station_id, type, value, recorded_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLatestReadingByType :one
SELECT * FROM readings
WHERE station_id = $1 AND type = $2
ORDER BY recorded_at DESC
LIMIT 1;

-- name: GetLatestReadings :many
SELECT DISTINCT ON (type) *
FROM readings
WHERE station_id = $1
ORDER BY type, recorded_at DESC;

-- name: ListReadingsByStation :many
SELECT * FROM readings
WHERE station_id = $1
ORDER BY recorded_at DESC
LIMIT $2;

-- name: ListReadingsByStationAndType :many
SELECT * FROM readings
WHERE station_id = $1 AND type = $2
ORDER BY recorded_at DESC
LIMIT $3;
