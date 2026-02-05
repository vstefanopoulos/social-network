-- Enable extensions
CREATE EXTENSION IF NOT EXISTS citext;

------------------------------------------
-- Master Index
------------------------------------------
CREATE TYPE content_type AS ENUM ('post','comment','event');

CREATE TABLE IF NOT EXISTS master_index (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    content_type content_type NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_master_type ON master_index(content_type);
CREATE INDEX idx_master_index_id_type ON master_index(id, content_type);

------------------------------------------
-- Posts
------------------------------------------
CREATE TYPE intended_audience AS ENUM ('everyone','followers','selected','group');

CREATE TABLE IF NOT EXISTS posts (
    id BIGINT PRIMARY KEY REFERENCES master_index(id) ON DELETE CASCADE,
    post_body TEXT NOT NULL,
    creator_id BIGINT NOT NULL, -- in user service
    group_id BIGINT, -- in user service, null for user posts
    audience intended_audience NOT NULL DEFAULT 'everyone',
    comments_count INT DEFAULT 0 NOT NULL,
    reactions_count INT DEFAULT 0 NOT NULL,
    last_commented_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
); 

CREATE INDEX idx_posts_creator ON posts(creator_id);
CREATE INDEX idx_posts_group ON posts(group_id);
CREATE INDEX idx_posts_audience_created ON posts(audience, created_at DESC);
CREATE INDEX idx_posts_creator_id_created_at ON posts(creator_id, created_at DESC);
CREATE INDEX idx_posts_group_id_created_at   ON posts(group_id, created_at DESC);
CREATE INDEX idx_posts_created_at_desc       ON posts(created_at DESC);
CREATE INDEX idx_posts_deleted_at_null       ON posts(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_active ON posts(id) WHERE deleted_at IS NULL;
CREATE INDEX idx_posts_personal_feed ON posts(created_at DESC) WHERE deleted_at IS NULL;

------------------------------------------
-- Post_audience (for 'selected' audience)
------------------------------------------
CREATE TABLE IF NOT EXISTS post_audience (
    post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    allowed_user_id BIGINT, -- in user service
    PRIMARY KEY (post_id, allowed_user_id)
);
CREATE INDEX idx_post_audience_post_id  ON post_audience(post_id);
CREATE INDEX idx_post_audience_user_id  ON post_audience(allowed_user_id);
CREATE UNIQUE INDEX uniq_post_user_audience ON post_audience(post_id, allowed_user_id);

------------------------------------------
-- Comments
------------------------------------------
CREATE TABLE IF NOT EXISTS comments (
    id BIGINT PRIMARY KEY REFERENCES master_index(id) ON DELETE CASCADE,
    comment_creator_id BIGINT NOT NULL, -- in users service
    parent_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    comment_body TEXT NOT NULL,
    reactions_count INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_comments_parent_created ON comments(parent_id, created_at DESC);
CREATE INDEX idx_comments_creator ON comments(comment_creator_id);
CREATE INDEX idx_comments_active ON comments(parent_id) WHERE deleted_at IS NULL;


------------------------------------------
-- Events
------------------------------------------
CREATE TABLE IF NOT EXISTS events (
    id BIGINT PRIMARY KEY REFERENCES master_index(id) ON DELETE CASCADE,
    event_title TEXT NOT NULL,
    event_body TEXT NOT NULL,
    event_creator_id BIGINT NOT NULL, -- in users service
    group_id BIGINT NOT NULL, -- in user service
    event_date DATE NOT NULL,
    going_count INT DEFAULT 0 NOT NULL,
    not_going_count INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_events_creator ON events(event_creator_id);
CREATE INDEX idx_events_date ON events(event_date);


------------------------------------------
-- Event responses
------------------------------------------
CREATE TABLE IF NOT EXISTS event_responses (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL, -- in users service
    going BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ux_event_user UNIQUE (event_id, user_id)
);

CREATE INDEX idx_event_responses_event ON event_responses(event_id);

------------------------------------------
-- Images
------------------------------------------
CREATE TABLE images (
    parent_id BIGINT PRIMARY KEY
        REFERENCES master_index(id) ON DELETE CASCADE,

    id  BIGINT NOT NULL,  
    sort_order INT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_images_active
ON images(parent_id)
WHERE deleted_at IS NULL;


------------------------------------------
-- Reactions (likes only)
------------------------------------------
CREATE TABLE IF NOT EXISTS reactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    content_id BIGINT NOT NULL REFERENCES master_index(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL, -- in users service
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT unique_user_reaction_per_content UNIQUE (user_id, content_id)
);


CREATE INDEX idx_reactions_user ON reactions(user_id);
CREATE INDEX idx_reactions_content_id ON reactions(content_id);
CREATE INDEX idx_reactions_active ON reactions(content_id) WHERE deleted_at IS NULL;



------------------------------------------
-- Trigger to auto insert to master_index
------------------------------------------
CREATE OR REPLACE FUNCTION add_to_master_index()
RETURNS TRIGGER AS $$
DECLARE
    new_id BIGINT;
    ctype content_type;
BEGIN
    -- Safely convert text argument to enum
    CASE TG_ARGV[0]
        WHEN 'post' THEN ctype := 'post'::content_type;
        WHEN 'comment' THEN ctype := 'comment'::content_type;
        WHEN 'event' THEN ctype := 'event'::content_type;
        ELSE RAISE EXCEPTION 'Unknown content_type: %', TG_ARGV[0];
    END CASE;
    
    INSERT INTO master_index (content_type)
    VALUES (ctype)
    RETURNING id INTO new_id;
    NEW.id := new_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_before_insert_post
BEFORE INSERT ON posts
FOR EACH ROW
EXECUTE FUNCTION add_to_master_index('post');

CREATE TRIGGER trg_before_insert_comment
BEFORE INSERT ON comments
FOR EACH ROW
EXECUTE FUNCTION add_to_master_index('comment');

CREATE TRIGGER trg_before_insert_event
BEFORE INSERT ON events
FOR EACH ROW
EXECUTE FUNCTION add_to_master_index('event');

------------------------------------------
-- Trigger to auto-update updated_at timestamps
------------------------------------------
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_post_modtime
BEFORE UPDATE ON posts
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_update_comment_modtime
BEFORE UPDATE ON comments
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_update_event_modtime
BEFORE UPDATE ON events
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_update_event_responses_modtime
BEFORE UPDATE ON event_responses
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_update_reactions_modtime
BEFORE UPDATE ON reactions
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER trg_update_images_modtime
BEFORE UPDATE ON images
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

------------------------------------------
-- Trigger to maintain comments_count and last_commented_at
------------------------------------------
CREATE OR REPLACE FUNCTION update_post_comments_count()
RETURNS TRIGGER AS $$
BEGIN
    -- New comment
    IF TG_OP = 'INSERT' THEN
        UPDATE posts
        SET comments_count = comments_count + 1,
            last_commented_at = NEW.created_at
        WHERE id = NEW.parent_id;

    -- Soft delete
    ELSIF TG_OP = 'UPDATE'
      AND OLD.deleted_at IS NULL
      AND NEW.deleted_at IS NOT NULL THEN

        UPDATE posts
        SET comments_count = GREATEST(comments_count - 1, 0),
            last_commented_at = (
                SELECT MAX(created_at)
                FROM comments
                WHERE parent_id = OLD.parent_id
                  AND deleted_at IS NULL
            )
        WHERE id = OLD.parent_id;

    -- Restore comment
    ELSIF TG_OP = 'UPDATE'
      AND OLD.deleted_at IS NOT NULL
      AND NEW.deleted_at IS NULL THEN

        UPDATE posts
        SET comments_count = comments_count + 1,
            last_commented_at = GREATEST(
                COALESCE(last_commented_at, NEW.created_at),
                NEW.created_at
            )
        WHERE id = NEW.parent_id;

    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;



CREATE TRIGGER trg_comments_insert
AFTER INSERT ON comments
FOR EACH ROW
EXECUTE FUNCTION update_post_comments_count();

CREATE TRIGGER trg_comments_update
AFTER UPDATE ON comments
FOR EACH ROW
EXECUTE FUNCTION update_post_comments_count();


------------------------------------------
-- Trigger to maintain reactions_count
------------------------------------------
CREATE OR REPLACE FUNCTION update_reactions_count()
RETURNS TRIGGER AS $$
DECLARE
    cid BIGINT;
    ctype content_type;
    delta INT;
BEGIN
    IF TG_OP = 'INSERT' THEN
        cid := NEW.content_id;
        delta := 1;
    ELSIF TG_OP = 'DELETE' THEN
        cid := OLD.content_id;
        delta := -1;
    ELSE
        RETURN NULL;
    END IF;

    SELECT content_type
    INTO ctype
    FROM master_index
    WHERE id = cid;

    IF ctype = 'post' THEN
        UPDATE posts
        SET reactions_count = GREATEST(reactions_count + delta, 0)
        WHERE id = cid;

    ELSIF ctype = 'comment' THEN
        UPDATE comments
        SET reactions_count = GREATEST(reactions_count + delta, 0)
        WHERE id = cid;
    END IF;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;


-- Attach the triggers
CREATE TRIGGER trg_reactions_insert
AFTER INSERT ON reactions
FOR EACH ROW
EXECUTE FUNCTION update_reactions_count();

CREATE TRIGGER trg_reactions_delete
AFTER DELETE ON reactions
FOR EACH ROW
EXECUTE FUNCTION update_reactions_count();


------------------------------------------
-- Soft delete cascade for posts, comments, events
------------------------------------------
CREATE OR REPLACE FUNCTION cascade_soft_delete_unified()
RETURNS TRIGGER AS $$
DECLARE
    _type TEXT;
BEGIN
    SELECT content_type INTO _type
    FROM master_index
    WHERE id = OLD.id;

    IF _type IN ('POST', 'COMMENT', 'EVENT') THEN
        UPDATE images
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE parent_id = OLD.id AND deleted_at IS NULL;
    END IF;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_soft_delete_post
AFTER UPDATE ON posts
FOR EACH ROW
WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
EXECUTE FUNCTION cascade_soft_delete_unified();

CREATE TRIGGER trg_soft_delete_comment
AFTER UPDATE ON comments
FOR EACH ROW
WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
EXECUTE FUNCTION cascade_soft_delete_unified();

CREATE TRIGGER trg_soft_delete_event
AFTER UPDATE ON events
FOR EACH ROW
WHEN (NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL)
EXECUTE FUNCTION cascade_soft_delete_unified();


------------------------------------------
-- Event responses count trigger
------------------------------------------
CREATE OR REPLACE FUNCTION update_event_response_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        IF NEW.going THEN
            UPDATE events SET going_count = going_count + 1 WHERE id = NEW.event_id;
        ELSE
            UPDATE events SET not_going_count = not_going_count + 1 WHERE id = NEW.event_id;
        END IF;

    ELSIF (TG_OP = 'UPDATE') THEN
        IF OLD.going <> NEW.going THEN
            IF NEW.going THEN
                UPDATE events
                SET going_count = going_count + 1,
                    not_going_count = not_going_count - 1
                WHERE id = NEW.event_id;
            ELSE
                UPDATE events
                SET going_count = going_count - 1,
                    not_going_count = not_going_count + 1
                WHERE id = NEW.event_id;
            END IF;
        END IF;

        -- Handle soft-delete / restore
        IF OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL THEN
            IF OLD.going THEN
                UPDATE events SET going_count = going_count - 1 WHERE id = NEW.event_id;
            ELSE
                UPDATE events SET not_going_count = not_going_count - 1 WHERE id = NEW.event_id;
            END IF;
        ELSIF OLD.deleted_at IS NOT NULL AND NEW.deleted_at IS NULL THEN
            IF OLD.going THEN
                UPDATE events SET going_count = going_count + 1 WHERE id = NEW.event_id;
            ELSE
                UPDATE events SET not_going_count = not_going_count + 1 WHERE id = NEW.event_id;
            END IF;
        END IF;

    ELSIF (TG_OP = 'DELETE') THEN
        IF OLD.deleted_at IS NULL THEN
            IF OLD.going THEN
                UPDATE events SET going_count = going_count - 1 WHERE id = OLD.event_id;
            ELSE
                UPDATE events SET not_going_count = not_going_count - 1 WHERE id = OLD.event_id;
            END IF;
        END IF;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_event_responses_counts_insert
AFTER INSERT ON event_responses
FOR EACH ROW
EXECUTE FUNCTION update_event_response_counts();

CREATE TRIGGER trg_event_responses_counts_update
AFTER UPDATE ON event_responses
FOR EACH ROW
EXECUTE FUNCTION update_event_response_counts();

CREATE TRIGGER trg_event_responses_counts_delete
AFTER DELETE ON event_responses
FOR EACH ROW
EXECUTE FUNCTION update_event_response_counts();

------------------------------------------
-- Images sort_order trigger
------------------------------------------
-- CREATE OR REPLACE FUNCTION set_next_sort_order()
-- RETURNS TRIGGER AS $$
-- DECLARE
--     max_order INT;
-- BEGIN
--     IF NEW.sort_order IS NULL THEN
--         SELECT COALESCE(MAX(sort_order),0) 
--         INTO max_order
--         FROM images
--         WHERE parent_id = NEW.parent_id;

--         NEW.sort_order := max_order + 1;
--     END IF;

--     RETURN NEW;
-- END;
-- $$ LANGUAGE plpgsql;

-- CREATE TRIGGER trg_set_sort_order
-- BEFORE INSERT ON images
-- FOR EACH ROW
-- EXECUTE FUNCTION set_next_sort_order();
