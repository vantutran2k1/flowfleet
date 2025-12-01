ALTER TABLE drivers
    ADD COLUMN email TEXT NOT NULL UNIQUE DEFAULT 'placeholder@example.com',
    ADD COLUMN password_hash TEXT NOT NULL DEFAULT '';

ALTER TABLE drivers ALTER COLUMN email DROP DEFAULT;
ALTER TABLE drivers ALTER COLUMN password_hash DROP DEFAULT;

CREATE INDEX idx_drivers_email ON drivers(email);