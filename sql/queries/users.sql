-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING id, created_at, updated_at, email, is_chirpy_red;

-- name: ResetUsers :exec
DELETE FROM users;

-- name: GetSingleUser :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateSingleUser :one
UPDATE users 
SET updated_at = NOW(), email = $2, hashed_password = $3
WHERE id = $1 
RETURNING id, created_at, updated_at, email, is_chirpy_red;


-- name: UpgradeSingleUserToChirpyRed :exec
UPDATE users 
SET updated_at = NOW(), is_chirpy_red = TRUE 
WHERE id = $1;
