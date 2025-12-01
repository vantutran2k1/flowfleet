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

-- name: RejectOrderAssignment :exec
UPDATE orders
SET driver_id = NULL,
    status = 'pending',
    updated_at = NOW()
WHERE id = $1 AND driver_id = $2 AND status = 'assigned';

-- name: ConfirmOrderAcceptance :exec
UPDATE drivers
SET status = 'en_route', updated_at = NOW()
WHERE id = $1;

-- name: MarkOrderArrived :execrows
UPDATE orders
SET status = 'arrived', updated_at = NOW()
WHERE id = $1 AND driver_id = $2 AND status = 'assigned';

-- name: MarkOrderPickedUp :execrows
UPDATE orders
SET status = 'picked_up', updated_at = NOW()
WHERE id = $1 AND driver_id = $2 AND status = 'arrived';

-- name: MarkOrderDelivered :execrows
UPDATE orders
SET status = 'delivered', updated_at = NOW()
WHERE id = $1 AND driver_id = $2 AND status = 'picked_up';