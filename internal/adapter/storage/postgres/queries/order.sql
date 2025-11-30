-- name: CreateOrder :one
INSERT INTO orders (fleet_id, amount_cents, status, pickup_location, dropoff_location)
VALUES ($1, $2, 'pending', ST_SetSRID(ST_MakePoint($3, $4), 4326), ST_SetSRID(ST_MakePoint($5, $6), 4326))
RETURNING id, created_at;

-- name: AssignDriverToOrder :exec
UPDATE orders
SET driver_id = $1, status = 'assigned', updated_at = NOW()
WHERE id = $2;

-- name: SetDriverStatus :exec
UPDATE drivers
SET status = $2, updated_at = NOW()
WHERE id = $1;