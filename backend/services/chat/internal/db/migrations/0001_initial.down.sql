-- Drop triggers
DROP TRIGGER IF EXISTS trg_update_group_conversations_modtime ON group_conversations;
DROP TRIGGER IF EXISTS trg_update_group_messages_modtime ON group_messages;
-- DROP TRIGGER IF EXISTS trg_update_private_conversations_modtime ON private_conversations;
DROP TRIGGER IF EXISTS trg_update_private_messages_modtime ON private_messages;
DROP TRIGGER IF EXISTS trg_touch_group_conversation ON private_messages;
DROP TRIGGER IF EXISTS trg_touch_private_conversation ON private_messages;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_timestamp();
DROP FUNCTION IF EXISTS touch_group_conversation();
DROP FUNCTION IF EXISTS touch_private_conversation();

-- Drop tables
DROP TABLE IF EXISTS group_messages;
DROP TABLE IF EXISTS private_messages;
DROP TABLE IF EXISTS group_conversations;
DROP TABLE IF EXISTS private_conversations;
