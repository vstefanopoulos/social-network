
-----------------------------------------
-- Trigger to auto-update updated_at timestamps
-----------------------------------------

-- Single trigger function for updated_at
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW IS DISTINCT FROM OLD THEN
        NEW.updated_at = CURRENT_TIMESTAMP;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach to multiple tables
CREATE TRIGGER trg_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_auth_user_updated_at
BEFORE UPDATE ON auth_user
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_groups_updated_at
BEFORE UPDATE ON groups
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_group_members_updated_at
BEFORE UPDATE ON group_members
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_group_join_requests_updated_at
BEFORE UPDATE ON group_join_requests
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_group_invites_updated_at
BEFORE UPDATE ON group_invites
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_follow_requests_updated_at
BEFORE UPDATE ON follow_requests
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();