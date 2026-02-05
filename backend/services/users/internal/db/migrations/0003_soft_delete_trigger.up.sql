-----------------------------------------
-- Soft delete cascade for users
-----------------------------------------
CREATE OR REPLACE FUNCTION soft_delete_user_cascade()
RETURNS TRIGGER AS $$
BEGIN
    -- Block soft-deleting a user who owns active groups
    IF EXISTS (
        SELECT 1 FROM groups
        WHERE group_owner = OLD.id
          AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Group owner cannot be deleted. Transfer ownership first.';
    END IF;


    -- Hard-delete follows (CASCADE handles this automatically)
    DELETE FROM follows
    WHERE follower_id = OLD.id OR following_id = OLD.id;

    -- Hard-delete follow requests (CASCADE handles this automatically)
    DELETE FROM follow_requests
    WHERE requester_id = OLD.id OR target_id = OLD.id;

    -- Soft-delete group memberships (preserve history)
    UPDATE group_members
    SET deleted_at = CURRENT_TIMESTAMP
    WHERE user_id = OLD.id AND deleted_at IS NULL;

    -- Soft-delete group join requests (preserve history)
    UPDATE group_join_requests
    SET deleted_at = CURRENT_TIMESTAMP
    WHERE user_id = OLD.id AND deleted_at IS NULL;

    -- Soft-delete group invites (preserve history)
    UPDATE group_invites
    SET deleted_at = CURRENT_TIMESTAMP
    WHERE (sender_id = OLD.id OR receiver_id = OLD.id) 
    AND deleted_at IS NULL;

    -- Handle owned groups
    UPDATE groups
    SET deleted_at = CURRENT_TIMESTAMP
    WHERE group_owner = OLD.id AND deleted_at IS NULL;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_soft_delete_user
BEFORE UPDATE ON users
FOR EACH ROW
WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
EXECUTE FUNCTION soft_delete_user_cascade();