-- name: CreateNotification :one
INSERT INTO notifications (
    user_id,
    notif_type,
    source_service,
    source_entity_id,
    needs_action,
    acted,
    payload,
    created_at,
    expires_at,
    count
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, NOW(), $8, $9
) RETURNING id, user_id, notif_type, source_service, source_entity_id, seen, needs_action, acted, payload, created_at, expires_at, deleted_at, count;

-- name: GetNotificationByID :one
SELECT id, user_id, notif_type, source_service, source_entity_id, seen, needs_action, acted, payload, created_at, expires_at, deleted_at, count
FROM notifications
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserNotifications :many
SELECT id, user_id, notif_type, source_service, source_entity_id, seen, needs_action, acted, payload, created_at, expires_at, deleted_at, count
FROM notifications
WHERE user_id = $1
  AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserNotificationsCount :one
SELECT COUNT(*)
FROM notifications
WHERE user_id = $1 AND deleted_at IS NULL;

-- name: GetUserUnreadNotificationsCount :one
SELECT COUNT(*)
FROM notifications
WHERE user_id = $1 AND seen = false AND deleted_at IS NULL;

-- name: MarkNotificationAsRead :exec
UPDATE notifications SET seen = true WHERE id = $1 AND user_id = $2;

-- name: MarkNotificationAsActed :exec
UPDATE notifications SET acted = true WHERE id = $1 AND user_id = $2;

-- name: MarkAllAsRead :exec
UPDATE notifications SET seen = true WHERE user_id = $1 AND seen = false;

-- name: DeleteNotification :exec
UPDATE notifications SET deleted_at = NOW() WHERE id = $1 AND user_id = $2;

-- name: UpdateNotificationCount :exec
UPDATE notifications SET count = $1 WHERE id = $2 AND user_id = $3;

-- name: GetUnreadNotificationByTypeAndEntity :one
SELECT id, user_id, notif_type, source_service, source_entity_id, seen, needs_action, acted, payload, created_at, expires_at, deleted_at, count
FROM notifications
WHERE user_id = $1 AND notif_type = $2 AND source_entity_id = $3 AND seen = false AND deleted_at IS NULL
LIMIT 1;

-- name: GetNotificationByTypeAndEntity :one
SELECT id, user_id, notif_type, source_service, source_entity_id, seen, needs_action, acted, payload, created_at, expires_at, deleted_at, count
FROM notifications
WHERE user_id = $1 AND notif_type = $2 AND source_entity_id = $3 AND deleted_at IS NULL
LIMIT 1;

-- name: CreateNotificationType :exec
INSERT INTO notification_types (notif_type, category, default_enabled)
VALUES ($1, $2, $3)
ON CONFLICT (notif_type) DO NOTHING;

-- name: GetNotificationType :one
SELECT notif_type, category, default_enabled
FROM notification_types
WHERE notif_type = $1;
