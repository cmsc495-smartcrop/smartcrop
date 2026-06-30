-- name: CreateStation :one
INSERT INTO stations (id, name, location)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetStation :one
SELECT * FROM stations
WHERE id = $1;

-- name: ListStations :many
SELECT * FROM stations
ORDER BY created_at DESC;
