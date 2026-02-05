CREATE TYPE file_visibility AS ENUM ('private', 'public');

CREATE TABLE IF NOT EXISTS files (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    filename      TEXT NOT NULL,              -- original filename
    mime_type     TEXT NOT NULL,
    size_bytes    BIGINT NOT NULL CHECK (size_bytes >= 0),

    bucket        TEXT NOT NULL,               -- uploads-originals
    object_key    TEXT NOT NULL,               -- path/key in MinIO

    visibility    file_visibility NOT NULL,
    status       TEXT NOT NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (bucket, object_key)
);

CREATE TABLE IF NOT EXISTS file_variants (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,

    file_id       BIGINT NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    mime_type     TEXT NOT NULL,
    size_bytes    BIGINT,

    variant       TEXT NOT NULL,        -- thumb, small, medium, large
    bucket        TEXT NOT NULL,
    object_key    TEXT NOT NULL,

    status       TEXT NOT NULL,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (file_id, variant),
    UNIQUE (bucket, object_key)
);

CREATE INDEX IF NOT EXISTS idx_file_variants_file_id 
    ON file_variants(file_id);




-- Auto-update timestamp
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_file_updated_at
BEFORE UPDATE ON files
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

CREATE TRIGGER update_variant_updated_at
BEFORE UPDATE ON file_variants
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();

-- Trigger to set variant of a file to processing when file_id status is set to complete.
CREATE OR REPLACE FUNCTION set_variants_processing()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'complete' AND OLD.status IS DISTINCT FROM NEW.status THEN
        UPDATE file_variants
        SET status = 'processing'
        WHERE file_id = NEW.id
            AND status <> 'complete';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER files_status_complete_trigger
AFTER UPDATE OF status ON files
FOR EACH ROW
WHEN (NEW.status = 'complete')   -- optional extra safety
EXECUTE FUNCTION set_variants_processing();

