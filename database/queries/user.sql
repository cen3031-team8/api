-- name: GetAllUsers :many
SELECT * FROM users;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING user_id, username, email, password_hash, created_at;

-- name: GetUserByUsername :one
SELECT user_id, username, email, password_hash, created_at FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT user_id, username, email, password_hash, created_at FROM users
WHERE user_id = $1;

-- name: GetUserInventory :many
SELECT user_id, item, quantity FROM user_inventory
WHERE user_id = $1;

-- name: GetUserInventoryItem :one
SELECT user_id, item, quantity FROM user_inventory
WHERE user_id = $1 AND item = $2;

-- name: AddOrUpdateInventoryItem :exec
INSERT INTO user_inventory (user_id, item, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, item) DO UPDATE SET
  quantity = quantity + $3;

-- name: UpdateInventoryItem :exec
UPDATE user_inventory
SET quantity = $3
WHERE user_id = $1 AND item = $2;
