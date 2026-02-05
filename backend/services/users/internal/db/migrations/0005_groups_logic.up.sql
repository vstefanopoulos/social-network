-----------------------------------------
-- Trigger to efficiently update members count for group
-----------------------------------------
CREATE OR REPLACE FUNCTION update_group_members_count()
RETURNS TRIGGER AS $$
BEGIN
    -- Member added (INSERT with no deleted_at)
    IF TG_OP = 'INSERT' AND NEW.deleted_at IS NULL THEN
        UPDATE groups
        SET members_count = members_count + 1
        WHERE id = NEW.group_id;
        RETURN NEW;
    END IF;

    -- Member restored from soft-delete
    IF TG_OP = 'UPDATE' AND OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
        UPDATE groups
        SET members_count = members_count + 1
        WHERE id = NEW.group_id;
        RETURN NEW;
    END IF;

    -- Member soft-deleted
    IF TG_OP = 'UPDATE' AND OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
        UPDATE groups
        SET members_count = members_count - 1
        WHERE id = NEW.group_id;
        RETURN NEW;
    END IF;

    -- Member hard-deleted 
    IF TG_OP = 'DELETE' AND OLD.deleted_at IS NULL THEN
        UPDATE groups
        SET members_count = members_count - 1
        WHERE id = OLD.group_id;
        RETURN OLD;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger for INSERT
CREATE TRIGGER trg_group_members_count_insert
AFTER INSERT ON group_members
FOR EACH ROW
EXECUTE FUNCTION update_group_members_count();

-- Trigger for UPDATE (soft-delete/restore)
CREATE TRIGGER trg_group_members_count_update
AFTER UPDATE ON group_members
FOR EACH ROW
EXECUTE FUNCTION update_group_members_count();

-- Trigger for DELETE (hard delete)
CREATE TRIGGER trg_group_members_count_delete
AFTER DELETE ON group_members
FOR EACH ROW
EXECUTE FUNCTION update_group_members_count();


-----------------------------------------
-- Trigger to add user as group member when join request accepted
-----------------------------------------
CREATE OR REPLACE FUNCTION add_group_member_on_join_accept()
RETURNS TRIGGER AS $$
BEGIN
    -- Add member, or restore if previously soft-deleted
    INSERT INTO group_members (group_id, user_id, role, joined_at)
    VALUES (NEW.group_id, NEW.user_id, 'member', CURRENT_TIMESTAMP)
    ON CONFLICT (group_id, user_id) DO UPDATE
    SET deleted_at = NULL,
        joined_at = CURRENT_TIMESTAMP
    WHERE group_members.deleted_at IS NOT NULL;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_add_group_member_on_join_accept
AFTER UPDATE ON group_join_requests
FOR EACH ROW
WHEN (NEW.status = 'accepted' AND OLD.status IS DISTINCT FROM 'accepted')
EXECUTE FUNCTION add_group_member_on_join_accept();


-----------------------------------------
-- Trigger to add user as group member when group invite accepted
-----------------------------------------
CREATE OR REPLACE FUNCTION add_group_member_on_invite_accept()
RETURNS TRIGGER AS $$
BEGIN
    -- Add member, or restore if previously soft-deleted
    INSERT INTO group_members (group_id, user_id, role, joined_at)
    VALUES (NEW.group_id, NEW.receiver_id, 'member', CURRENT_TIMESTAMP)
    ON CONFLICT (group_id, user_id) DO UPDATE
    SET deleted_at = NULL,
        joined_at = CURRENT_TIMESTAMP
    WHERE group_members.deleted_at IS NOT NULL;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_add_group_member_on_invite_accept
AFTER UPDATE ON group_invites
FOR EACH ROW
WHEN (NEW.status = 'accepted' AND OLD.status IS DISTINCT FROM 'accepted')
EXECUTE FUNCTION add_group_member_on_invite_accept();


-----------------------------------------
-- Trigger to add group owner as member on group creation
-----------------------------------------
CREATE OR REPLACE FUNCTION add_group_owner_as_member()
RETURNS TRIGGER AS $$
BEGIN
    -- Add the group owner as a member with 'owner' role
    INSERT INTO group_members (group_id, user_id, role, joined_at)
    VALUES (NEW.id, NEW.group_owner, 'owner', NEW.created_at)
    ON CONFLICT (group_id, user_id) DO NOTHING;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_add_group_owner_as_member
AFTER INSERT ON groups
FOR EACH ROW
EXECUTE FUNCTION add_group_owner_as_member();


-----------------------------------------
-- Trigger to prevent owner from leaving their own group
-----------------------------------------
CREATE OR REPLACE FUNCTION prevent_owner_leave()
RETURNS TRIGGER AS $$
BEGIN
    -- Prevent soft-deleting the owner
    IF TG_OP = 'UPDATE' 
       AND OLD.role = 'owner' 
       AND OLD.deleted_at IS NULL 
       AND NEW.deleted_at IS NOT NULL 
    THEN
        RAISE EXCEPTION 'Group owner cannot leave the group. Transfer ownership first.';
    END IF;

    -- Prevent hard-deleting the owner
    IF TG_OP = 'DELETE' AND OLD.role = 'owner' AND OLD.deleted_at IS NULL THEN
        RAISE EXCEPTION 'Group owner cannot be removed. Transfer ownership first.';
    END IF;

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    ELSE
        RETURN NEW;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_owner_leave
BEFORE UPDATE OR DELETE ON group_members
FOR EACH ROW
EXECUTE FUNCTION prevent_owner_leave();


-----------------------------------------
-- Trigger to prevent group owners from being deleted
-----------------------------------------
CREATE OR REPLACE FUNCTION prevent_delete_group_owner()
RETURNS TRIGGER AS $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM groups 
        WHERE group_owner = OLD.id 
        AND deleted_at IS NULL
    ) THEN
        RAISE EXCEPTION 'Cannot delete user who owns groups. Transfer ownership first.';
    END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_delete_group_owner
BEFORE DELETE ON users
FOR EACH ROW
EXECUTE FUNCTION prevent_delete_group_owner();