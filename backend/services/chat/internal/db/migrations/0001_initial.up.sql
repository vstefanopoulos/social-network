------------------------------------------------------------
-- 1. GROUP CONVERSATIONS
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS group_conversations (
    group_id BIGINT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_group_conversations_updated_at_active
ON group_conversations (updated_at DESC)
WHERE deleted_at IS NULL;

------------------------------------------------------------
-- 2. GROUP MESSAGES
------------------------------------------------------------

CREATE TABLE IF NOT EXISTS group_messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES group_conversations(group_id),
    sender_id BIGINT NOT NULL,
    message_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_group_messages_sender ON group_messages(sender_id);
CREATE INDEX idx_group_messages_created_at ON group_messages(created_at);
CREATE INDEX idx_group_messages_group_id_id
    ON group_messages(group_id, id);



------------------------------------------------------------
-- 4. PRIVATE CONVERSATIONS
------------------------------------------------------------
CREATE TABLE IF NOT EXISTS private_conversations (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_a BIGINT NOT NULL,
    user_b BIGINT NOT NULL,
    last_read_message_id_a BIGINT,
    last_read_message_id_b BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CHECK (user_a < user_b)
);


CREATE UNIQUE INDEX uq_private_conversation_users
ON private_conversations(user_a, user_b);

CREATE INDEX idx_private_conversations_updated_at_active
ON private_conversations (updated_at DESC)
WHERE deleted_at IS NULL;



------------------------------------------------------------
-- 5. PRIVATE MESSAGES
------------------------------------------------------------

CREATE TABLE IF NOT EXISTS private_messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES private_conversations(id),
    sender_id BIGINT NOT NULL,
    message_text TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_private_messages_conversation ON private_messages(conversation_id);
CREATE INDEX idx_private_messages_sender ON private_messages(sender_id);
CREATE INDEX idx_private_messages_created_at ON private_messages(created_at);
CREATE INDEX idx_private_messages_conversation_id_id
    ON private_messages(conversation_id, id);


--------------------------------------------------------------
-- 6. ALTER PRIVATE CONVERSATIONS
--------------------------------------------------------------

ALTER TABLE private_conversations
    ADD CONSTRAINT private_conversations_last_read_message_id_a_id_fkey
        FOREIGN KEY (last_read_message_id_a)
        REFERENCES private_messages(id)
        ON DELETE SET NULL;

ALTER TABLE private_conversations
    ADD CONSTRAINT private_conversations_last_read_message_id_b_id_fkey
        FOREIGN KEY (last_read_message_id_b)
        REFERENCES private_messages(id)
        ON DELETE SET NULL;


------------------------------------------------------------
-- 7. Triggers: Auto-update updated_at
------------------------------------------------------------

-- UPDATE TIMESTAMP
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_group_conversations_modtime
BEFORE UPDATE ON group_conversations
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_update_group_messages_modtime
BEFORE UPDATE ON group_messages
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

-- CREATE TRIGGER trg_update_private_conversations_modtime
-- BEFORE UPDATE ON private_conversations
-- FOR EACH ROW EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_update_private_messages_modtime
BEFORE UPDATE ON private_messages
FOR EACH ROW EXECUTE FUNCTION update_timestamp();


-- UPDATE GROUP CONVERSATION updated_at WHEN NEW MESSAGE
CREATE OR REPLACE FUNCTION touch_group_conversation()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE group_conversations
    SET updated_at = CURRENT_TIMESTAMP
    WHERE group_id = NEW.group_id
        AND deleted_at IS NULL;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_touch_group_conversation
AFTER INSERT ON group_messages
FOR EACH ROW
EXECUTE FUNCTION touch_group_conversation();

-- UPDATE PRIVATE CONVERSATION updated_at WHEN NEW MESSAGE
CREATE OR REPLACE FUNCTION touch_private_conversation()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE private_conversations
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.conversation_id
        AND deleted_at IS NULL;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_touch_private_conversation
AFTER INSERT ON private_messages
FOR EACH ROW
EXECUTE FUNCTION touch_private_conversation();
