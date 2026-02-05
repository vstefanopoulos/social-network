-- Enable extensions
CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Case-insensitive collation
CREATE COLLATION IF NOT EXISTS case_insensitive_ai (
  provider = icu,
  locale = 'und-u-ks-level1',
  deterministic = false
);

-----------------------------------------
-- Users table
-----------------------------------------
CREATE TYPE user_status AS ENUM ('active', 'banned', 'deleted');

CREATE TABLE IF NOT EXISTS users (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username CITEXT COLLATE case_insensitive_ai NOT NULL,
    first_name VARCHAR(255) COLLATE "C" NOT NULL,
    last_name VARCHAR(255) COLLATE "C" NOT NULL,
    date_of_birth DATE NOT NULL,
    avatar_id BIGINT NOT NULL,
    about_me TEXT NOT NULL, 
    profile_public BOOLEAN NOT NULL DEFAULT TRUE,
    current_status user_status NOT NULL DEFAULT 'active',
    ban_ends_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_users_id ON users(id);
CREATE INDEX idx_users_status ON users(current_status);

CREATE INDEX users_username_trgm_idx
    ON users USING GIN (username gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX users_first_name_trgm_idx
    ON users USING GIN (first_name gin_trgm_ops)
    WHERE deleted_at IS NULL;

CREATE INDEX users_last_name_trgm_idx
    ON users USING GIN (last_name gin_trgm_ops)
    WHERE deleted_at IS NULL;



-----------------------------------------
-- Auth table (one-to-one with users)
-----------------------------------------
CREATE TABLE IF NOT EXISTS auth_user (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    email CITEXT COLLATE case_insensitive_ai UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ
);


-----------------------------------------
-- Follows
-----------------------------------------
CREATE TABLE IF NOT EXISTS follows (
    follower_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    following_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    PRIMARY KEY (follower_id, following_id),
    CONSTRAINT no_self_follow CHECK (follower_id <> following_id)
);


CREATE INDEX idx_follows_follower ON follows(follower_id);
CREATE INDEX idx_follows_following ON follows(following_id);

CREATE INDEX IF NOT EXISTS idx_follows_active_follower
ON follows(follower_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_follows_active_following
ON follows(following_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_follows_active_pair
ON follows(follower_id, following_id)
WHERE deleted_at IS NULL;

-----------------------------------------
-- Follow requests
-----------------------------------------
CREATE TYPE follow_request_status AS ENUM ('pending','accepted','rejected');

CREATE TABLE IF NOT EXISTS follow_requests (
    requester_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    target_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    status follow_request_status NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (requester_id, target_id)
);

CREATE INDEX idx_follow_requests_target_status ON follow_requests(target_id, status);


-----------------------------------------
-- Groups
-----------------------------------------
CREATE TABLE IF NOT EXISTS groups (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    group_owner BIGINT NOT NULL REFERENCES users(id) ON DELETE NO ACTION,
    group_title TEXT NOT NULL,
    group_description TEXT NOT NULL,
    group_image_id BIGINT NOT NULL, 
    members_count INT NOT NULL DEFAULT 0, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

-- Unique constraint on active, case-insensitive titles
CREATE UNIQUE INDEX groups_unique_title_active
ON groups (LOWER(group_title))
WHERE deleted_at IS NULL;

CREATE INDEX idx_groups_owner ON groups(group_owner);

CREATE INDEX idx_groups_title_trgm
    ON groups USING gin (group_title gin_trgm_ops);

CREATE INDEX idx_groups_description_trgm
    ON groups USING gin (group_description gin_trgm_ops);


-----------------------------------------
-- Group members
-----------------------------------------
CREATE TYPE group_role AS ENUM ('member','owner');

CREATE TABLE IF NOT EXISTS group_members (
    group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    role group_role DEFAULT 'member',
    joined_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (group_id, user_id)
);

-- Ensure exactly one owner per group
CREATE UNIQUE INDEX IF NOT EXISTS idx_group_one_owner
ON group_members(group_id)
WHERE role='owner' AND deleted_at IS NULL;

CREATE INDEX idx_group_members_user ON group_members(user_id);


-----------------------------------------
-- Group join requests
-----------------------------------------
CREATE TYPE join_request_status AS ENUM ('pending','accepted','rejected');

CREATE TABLE IF NOT EXISTS group_join_requests (
    group_id BIGINT REFERENCES groups(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    status join_request_status NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (group_id, user_id)
);

CREATE INDEX idx_group_join_requests_status ON group_join_requests(status);


-----------------------------------------
-- Group invites
-----------------------------------------
CREATE TYPE group_invite_status AS ENUM ('pending','accepted','declined','expired');

CREATE TABLE IF NOT EXISTS group_invites (
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status group_invite_status NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    PRIMARY KEY (group_id, receiver_id)
);

CREATE INDEX idx_group_invites_status ON group_invites(status);



