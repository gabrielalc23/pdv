-- name: GetUserByNormalizedEmail :one
SELECT
    u.id,
    u.email,
    u.email_normalized,
    u.display_name,
    u.status,
    u.email_verified_at,
    u.password_version,
    u.last_login_at,
    u.created_at,
    u.updated_at,
    p.password_hash,
    p.changed_at AS password_changed_at
FROM users AS u
JOIN user_passwords AS p ON p.user_id = u.id
WHERE u.email_normalized = sqlc.arg(email_normalized)
LIMIT 1;

-- name: GetUserForActionByNormalizedEmail :one
SELECT
    u.id,
    u.email,
    u.display_name,
    u.status,
    u.email_verified_at,
    EXISTS (
        SELECT 1
        FROM user_passwords AS p
        WHERE p.user_id = u.id
    ) AS has_password
FROM users AS u
WHERE u.email_normalized = sqlc.arg(email_normalized)
LIMIT 1;

-- name: GetUserByID :one
SELECT
    id,
    email,
    email_normalized,
    display_name,
    status,
    email_verified_at,
    password_version,
    last_login_at,
    created_at,
    updated_at
FROM users
WHERE id = sqlc.arg(user_id);

-- name: CreateUserWithPassword :one
WITH created_user AS (
    INSERT INTO users (
        email,
        email_normalized,
        display_name
    )
    VALUES (
        sqlc.arg(email),
        sqlc.arg(email_normalized),
        sqlc.arg(display_name)
    )
    RETURNING
        id,
        email,
        email_normalized,
        display_name,
        status,
        email_verified_at,
        password_version,
        last_login_at,
        created_at,
        updated_at
), created_password AS (
    INSERT INTO user_passwords (
        user_id,
        password_hash
    )
    SELECT
        id,
        sqlc.arg(password_hash)
    FROM created_user
    RETURNING user_id
)
SELECT
    u.id,
    u.email,
    u.email_normalized,
    u.display_name,
    u.status,
    u.email_verified_at,
    u.password_version,
    u.last_login_at,
    u.created_at,
    u.updated_at
FROM created_user AS u
JOIN created_password AS p ON p.user_id = u.id;

-- name: UpdateUserPassword :one
UPDATE user_passwords
SET
    password_hash = sqlc.arg(password_hash),
    changed_at = NOW()
WHERE user_id = sqlc.arg(user_id)
RETURNING
    user_id,
    changed_at,
    created_at,
    updated_at;

-- name: GetUserWithPasswordForUpdate :one
SELECT
    u.id,
    u.email,
    u.email_normalized,
    u.display_name,
    u.status,
    u.email_verified_at,
    u.password_version,
    p.password_hash,
    p.changed_at AS password_changed_at
FROM users AS u
JOIN user_passwords AS p ON p.user_id = u.id
WHERE u.id = sqlc.arg(user_id)
FOR UPDATE OF u, p;

-- name: UpdateUserPasswordAndIncrementVersion :one
WITH updated_password AS (
    UPDATE user_passwords
    SET
        password_hash = sqlc.arg(password_hash),
        changed_at = NOW()
    WHERE user_id = sqlc.arg(user_id)
    RETURNING user_id, changed_at
), updated_user AS (
    UPDATE users AS u
    SET password_version = u.password_version + 1
    FROM updated_password AS p
    WHERE u.id = p.user_id
    RETURNING
        u.id,
        (u.password_version - 1)::BIGINT AS previous_password_version,
        u.password_version AS new_password_version
)
SELECT
    u.id,
    u.previous_password_version,
    u.new_password_version,
    p.changed_at
FROM updated_user AS u
JOIN updated_password AS p ON p.user_id = u.id;

-- name: IncrementUserPasswordVersion :one
UPDATE users
SET password_version = password_version + 1
WHERE id = sqlc.arg(user_id)
RETURNING
    id,
    password_version,
    updated_at;

-- name: VerifyUserEmail :one
UPDATE users
SET email_verified_at = COALESCE(email_verified_at, NOW())
WHERE id = sqlc.arg(user_id)
RETURNING
    id,
    email_verified_at,
    updated_at;

-- name: UpdateUserLastLogin :one
UPDATE users
SET last_login_at = NOW()
WHERE id = sqlc.arg(user_id)
RETURNING id, last_login_at, updated_at;

-- name: UpdateUserDisplayName :one
UPDATE users
SET display_name = sqlc.arg(display_name)
WHERE id = sqlc.arg(user_id)
RETURNING id, email, display_name, email_verified_at, updated_at;

-- name: ListUserActiveMemberships :many
SELECT
    m.id AS membership_id,
    m.organization_id,
    m.status AS membership_status,
    m.authorization_version AS membership_authorization_version,
    m.default_store_id,
    m.joined_at,
    o.name AS organization_name,
    o.slug AS organization_slug,
    o.status AS organization_status,
    o.timezone AS organization_timezone,
    o.locale AS organization_locale,
    o.currency AS organization_currency,
    o.authorization_version AS organization_authorization_version,
    s.code AS default_store_code,
    s.name AS default_store_name,
    s.status AS default_store_status
FROM organization_memberships AS m
JOIN organizations AS o ON o.id = m.organization_id
LEFT JOIN stores AS s
  ON s.organization_id = m.organization_id
 AND s.id = m.default_store_id
WHERE m.user_id = sqlc.arg(user_id)
  AND m.status = 'ACTIVE'
  AND o.status = 'ACTIVE'
ORDER BY m.joined_at ASC, m.id ASC;

-- name: GetUserIdentityByNormalizedEmail :one
SELECT id, email, email_normalized, display_name, status, email_verified_at,
       password_version, last_login_at, created_at, updated_at
FROM users
WHERE email_normalized = sqlc.arg(email_normalized);
