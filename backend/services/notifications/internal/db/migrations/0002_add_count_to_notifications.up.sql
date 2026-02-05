-- Add count field to notifications table for aggregation functionality
ALTER TABLE notifications ADD COLUMN count INTEGER DEFAULT 1;

-- Update the index to include count for better query performance
DROP INDEX IF EXISTS idx_notifications_user_unread;
CREATE INDEX idx_notifications_user_unread
ON notifications (user_id, seen, created_at DESC)
WHERE deleted_at IS NULL;

DROP INDEX IF EXISTS idx_notifications_user_created;
CREATE INDEX idx_notifications_user_created
ON notifications (user_id, created_at DESC)
WHERE deleted_at IS NULL;