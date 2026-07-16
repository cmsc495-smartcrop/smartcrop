-- name: CreateStation :one
INSERT INTO stations (id, name, latitude, longitude)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpsertStation :one
INSERT INTO stations (id, name, latitude, longitude)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
RETURNING *;

-- name: GetStation :one
SELECT * FROM stations
WHERE id = $1;

-- name: ListStations :many
SELECT * FROM stations
ORDER BY created_at DESC;
