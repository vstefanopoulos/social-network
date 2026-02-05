-- Add new response notification types to ensure they exist in existing databases
-- These were added to the initial migration but for existing deployments we need this separate migration

INSERT INTO notification_types (notif_type, category, default_enabled)
VALUES
  ('follow_request_accepted', 'social',  TRUE),
  ('follow_request_rejected', 'social',  TRUE),
  ('group_invite_accepted', 'group',  TRUE),
  ('group_invite_rejected', 'group',  TRUE),
  ('group_join_request_accepted', 'group',  TRUE),
  ('group_join_request_rejected', 'group',  TRUE)
ON CONFLICT (notif_type) DO NOTHING;