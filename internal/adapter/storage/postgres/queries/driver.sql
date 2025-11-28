-- name: CreateDriver :one
INSERT INTO drivers (fleet_id, name, phone, status, current_location)
VALUES ($1, $2, $3, $4, ST_SetSRID(ST_MakePoint($5, $6), 4326))
RETURNING id, created_at;

-- name: GetDriver :one
SELECT id, fleet_id, name, phone, status, ST_AsText(current_location) as location
FROM drivers
WHERE id = $1 LIMIT 1;

-- name: ListDriversByFleet :many
SELECT id, name, phone, status
FROM drivers
WHERE fleet_id = $1
ORDER BY name;

-- name: FindNearestDrivers :many
SELECT id, name, status,
       ST_Distance(current_location, ST_SetSRID(ST_MakePoint($2, $3), 4326)) as dist_meters
FROM drivers
WHERE fleet_id = $1 
  AND status = 'idle'
ORDER BY current_location <-> ST_SetSRID(ST_MakePoint($2, $3), 4326)
LIMIT 10;