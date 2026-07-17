-- name: CreateAuditEvent :one
INSERT INTO security_audit_events (
    organization_id,
    store_id,
    actor_user_id,
    actor_membership_id,
    session_id,
    event_type,
    entity_type,
    entity_id,
    request_id,
    ip_address,
    user_agent,
    outcome,
    metadata
)
VALUES (
    sqlc.narg(organization_id),
    sqlc.narg(store_id),
    sqlc.narg(actor_user_id),
    sqlc.narg(actor_membership_id),
    sqlc.narg(session_id),
    sqlc.arg(event_type),
    sqlc.narg(entity_type),
    sqlc.narg(entity_id),
    sqlc.narg(request_id),
    sqlc.narg(ip_address),
    sqlc.narg(user_agent),
    sqlc.arg(outcome),
    sqlc.arg(metadata)
)
RETURNING
    id,
    organization_id,
    store_id,
    actor_user_id,
    actor_membership_id,
    session_id,
    event_type,
    entity_type,
    entity_id,
    request_id,
    ip_address,
    user_agent,
    outcome,
    metadata,
    occurred_at;

-- name: ListAuditEvents :many
SELECT
    id,
    organization_id,
    store_id,
    actor_user_id,
    actor_membership_id,
    session_id,
    event_type,
    entity_type,
    entity_id,
    request_id,
    ip_address,
    user_agent,
    outcome,
    metadata,
    occurred_at
FROM security_audit_events
WHERE organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(store_id) AS UUID) IS NULL
      OR store_id = CAST(sqlc.narg(store_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(actor_user_id) AS UUID) IS NULL
      OR actor_user_id = CAST(sqlc.narg(actor_user_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(actor_membership_id) AS UUID) IS NULL
      OR actor_membership_id = CAST(sqlc.narg(actor_membership_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(event_type) AS TEXT) IS NULL
      OR event_type = CAST(sqlc.narg(event_type) AS TEXT)
  )
  AND (
      CAST(sqlc.narg(outcome) AS audit_outcome) IS NULL
      OR outcome = CAST(sqlc.narg(outcome) AS audit_outcome)
  )
  AND (
      CAST(sqlc.narg(entity_type) AS TEXT) IS NULL
      OR entity_type = CAST(sqlc.narg(entity_type) AS TEXT)
  )
  AND (
      CAST(sqlc.narg(entity_id) AS UUID) IS NULL
      OR entity_id = CAST(sqlc.narg(entity_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(occurred_from) AS TIMESTAMPTZ) IS NULL
      OR occurred_at >= CAST(sqlc.narg(occurred_from) AS TIMESTAMPTZ)
  )
  AND (
      CAST(sqlc.narg(occurred_to) AS TIMESTAMPTZ) IS NULL
      OR occurred_at < CAST(sqlc.narg(occurred_to) AS TIMESTAMPTZ)
  )
ORDER BY occurred_at DESC, id DESC
LIMIT sqlc.arg(page_size)
OFFSET sqlc.arg(page_offset);

-- name: CountAuditEvents :one
SELECT COUNT(*)
FROM security_audit_events
WHERE organization_id = sqlc.arg(organization_id)
  AND (
      CAST(sqlc.narg(store_id) AS UUID) IS NULL
      OR store_id = CAST(sqlc.narg(store_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(actor_user_id) AS UUID) IS NULL
      OR actor_user_id = CAST(sqlc.narg(actor_user_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(actor_membership_id) AS UUID) IS NULL
      OR actor_membership_id = CAST(sqlc.narg(actor_membership_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(event_type) AS TEXT) IS NULL
      OR event_type = CAST(sqlc.narg(event_type) AS TEXT)
  )
  AND (
      CAST(sqlc.narg(outcome) AS audit_outcome) IS NULL
      OR outcome = CAST(sqlc.narg(outcome) AS audit_outcome)
  )
  AND (
      CAST(sqlc.narg(entity_type) AS TEXT) IS NULL
      OR entity_type = CAST(sqlc.narg(entity_type) AS TEXT)
  )
  AND (
      CAST(sqlc.narg(entity_id) AS UUID) IS NULL
      OR entity_id = CAST(sqlc.narg(entity_id) AS UUID)
  )
  AND (
      CAST(sqlc.narg(occurred_from) AS TIMESTAMPTZ) IS NULL
      OR occurred_at >= CAST(sqlc.narg(occurred_from) AS TIMESTAMPTZ)
  )
  AND (
      CAST(sqlc.narg(occurred_to) AS TIMESTAMPTZ) IS NULL
      OR occurred_at < CAST(sqlc.narg(occurred_to) AS TIMESTAMPTZ)
  );
