-- +goose Up
ALTER TYPE reading_type ADD VALUE 'wind_direction';

-- +goose Down
-- Postgres does not support removing enum values without rebuilding the type.
-- To rollback: recreate reading_type without wind_direction and cast existing data.
