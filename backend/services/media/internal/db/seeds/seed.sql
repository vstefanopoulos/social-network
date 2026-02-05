-- Clean up existing entries for IDs 1-10 to avoid conflicts
DELETE FROM file_variants WHERE file_id >= 1 AND file_id <= 10;
DELETE FROM files WHERE id >= 1 AND id <= 10;

-- Insert Files
INSERT INTO files (filename, mime_type, size_bytes, bucket, object_key, visibility, status)
OVERRIDING SYSTEM VALUE VALUES
('image1.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key1', 'public', 'complete'),
('image2.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key2', 'public', 'complete'),
('image3.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key3', 'public', 'pending'),
('image4.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key4', 'public', 'pending'),
('image5.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key5', 'public', 'processing'),
('image6.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key6', 'public', 'processing'),
('image7.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key7', 'public', 'failed'),
('image8.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key8', 'public', 'failed'),
('image9.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key9', 'public', 'complete'),
('image10.jpg', 'image/jpeg', 1024, 'uploads-originals', 'key10', 'public', 'complete');

-- Fix sequence if needed (optional, depends on usage)
-- SELECT setval(pg_get_serial_sequence('files', 'id'), (SELECT MAX(id) FROM files));

-- Insert Variants
-- Assuming we want at least a 'thumb' variant for testing GetImages which likely requests a variant.
INSERT INTO file_variants (file_id, mime_type, size_bytes, variant, bucket, object_key, status)
OVERRIDING SYSTEM VALUE VALUES
-- File 1: Complete
(1, 'image/webp', 512, 'thumb', 'uploads-variants', 'key1/thumb', 'complete'),

-- File 2: Complete
(2, 'image/webp', 512, 'thumb', 'uploads-variants', 'key2/thumb', 'complete'),

-- File 3: Pending
(3, 'image/webp', 512, 'thumb', 'uploads-variants', 'key3/thumb', 'pending'),

-- File 4: Pending
(4, 'image/webp', 512, 'thumb', 'uploads-variants', 'key4/thumb', 'pending'),

-- File 5: Processing
(5, 'image/webp', 512, 'thumb', 'uploads-variants', 'key5/thumb', 'complete'),

-- File 6: Processing
(6, 'image/webp', 512, 'thumb', 'uploads-variants', 'key6/thumb', 'complete'),

-- File 7: Failed
(7, 'image/webp', 512, 'thumb', 'uploads-variants', 'key7/thumb', 'failed'),

-- File 8: Failed
(8, 'image/webp', 512, 'thumb', 'uploads-variants', 'key8/thumb', 'failed'),

-- File 9: Complete
(9, 'image/webp', 512, 'thumb', 'uploads-variants', 'key9/thumb', 'complete'),

-- File 10: Complete
(10, 'image/webp', 512, 'thumb', 'uploads-variants', 'key10/thumb', 'complete');

-- Fix sequence for variants if needed
-- SELECT setval(pg_get_serial_sequence('file_variants', 'id'), (SELECT MAX(id) FROM file_variants));
