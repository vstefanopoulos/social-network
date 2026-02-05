-------------------------------------------------------
-- Development Seed Data
-- Safe for non-production environments only
-- Idempotent: safe to run multiple times
-- Uses ON CONFLICT DO NOTHING to skip existing records
-------------------------------------------------------

-------------------------------------------------------
-- Users and Auth Records in Transaction
-------------------------------------------------------
BEGIN TRANSACTION;

INSERT INTO users (username, first_name, last_name, date_of_birth, avatar_id, about_me, profile_public, current_status)
OVERRIDING SYSTEM VALUE
VALUES
('alice', 'Alice', 'Wonder', '1990-01-01', 0, 'Love nature and outdoor activities', TRUE, 'active'),
('bob', 'Bob', 'Builder', '1992-02-02', 0, 'Professional builder and contractor', FALSE, 'active'),
('charlie', 'Charlie', 'Day', '1991-03-03', 0, 'Gamer and tech enthusiast', TRUE, 'active'),
('diana', 'Diana', 'Prince', '1988-04-04', 0, 'Entrepreneur and business owner', TRUE, 'active'),
('eve', 'Eve', 'Hacker', '1995-05-05', 0, 'Security researcher', FALSE, 'active'),
('frank', 'Frank', 'Ocean', '1994-06-06', 0, 'Music producer and artist', TRUE, 'active'),
('grace', 'Grace', 'Hopper', '1985-07-07', 0, 'Software engineer and mentor', TRUE, 'active'),
('henry', 'Henry', 'Ford', '1986-08-08', 0, 'Automotive engineer', FALSE, 'active'),
('ivy', 'Ivy', 'Green', '1993-09-09', 0, 'Environmental activist', TRUE, 'active'),
('jack', 'Jack', 'Black', '1990-10-10', 0, 'Actor and musician', TRUE, 'active')
ON CONFLICT (id) DO NOTHING;

INSERT INTO auth_user (user_id, email, password_hash)
SELECT u.id,
       LOWER(u.username) || '@example.com' AS email,
       '+uFqO/5z/4l7Rw/bVTkULcce9TLoVz5ciXpBQyqSb4Q=' AS password_hash
FROM users u
LEFT JOIN auth_user a ON a.user_id = u.id
WHERE a.user_id IS NULL;

COMMIT;  

-------------------------------------------------------
-- Follow relationships (direct)
-------------------------------------------------------
INSERT INTO follows (follower_id, following_id)
VALUES
(1, 3), -- Alice → Charlie
(1, 4), -- Alice → Diana
(3, 1), -- Charlie → Alice (mutual)
(4, 3), -- Diana → Charlie
(6, 1), -- Frank → Alice
(9, 1), -- Ivy → Alice
(10, 3) -- Jack → Charlie
ON CONFLICT (follower_id, following_id) DO NOTHING;

-------------------------------------------------------
-- Follow requests (for private profiles)
-------------------------------------------------------
INSERT INTO follow_requests (requester_id, target_id, status)
VALUES
(1, 2, 'pending'), -- Alice → Bob (private)
(3, 2, 'accepted'), -- Charlie → Bob
(4, 5, 'pending'), -- Diana → Eve (private)
(7, 8, 'rejected') -- Grace → Henry (private)
ON CONFLICT (requester_id, target_id) DO NOTHING;

-------------------------------------------------------
-- Groups
-------------------------------------------------------
INSERT INTO groups (group_owner, group_title, group_description,group_image_id)
OVERRIDING SYSTEM VALUE
VALUES
(1, 'Nature Lovers', 'A group for nature enthusiasts',0),
(3, 'Gamers Unite', 'All about gaming',0),
(6, 'Music Fans', 'People who love music',0)
ON CONFLICT DO NOTHING;

-------------------------------------------------------
-- Group Members
-------------------------------------------------------
-- Group 1 (Nature Lovers)
INSERT INTO group_members (group_id, user_id, role)
VALUES
(1, 1, 'owner'),
(1, 3, 'member'),
(1, 4, 'member')
ON CONFLICT (group_id, user_id) DO NOTHING;

-- Group 2 (Gamers Unite)
INSERT INTO group_members (group_id, user_id, role)
VALUES
(2, 3, 'owner'),
(2, 6, 'member'),
(2, 7, 'member')
ON CONFLICT (group_id, user_id) DO NOTHING;

-- Group 3 (Music Fans)
INSERT INTO group_members (group_id, user_id, role)
VALUES
(3, 6, 'owner'),
(3, 1, 'member'),
(3, 9, 'member')
ON CONFLICT (group_id, user_id) DO NOTHING;

-------------------------------------------------------
-- Group Join Requests
-------------------------------------------------------
INSERT INTO group_join_requests (group_id, user_id, status)
VALUES
(1, 5, 'pending'),
(2, 1, 'rejected'),
(3, 10, 'accepted')
ON CONFLICT (group_id, user_id) DO NOTHING;

-------------------------------------------------------
-- Group Invites
-------------------------------------------------------
INSERT INTO group_invites (group_id, sender_id, receiver_id, status)
VALUES
(1, 1, 2, 'pending'),
(2, 3, 5, 'declined'),
(3, 6, 8, 'accepted')
ON CONFLICT (group_id, receiver_id) DO NOTHING;
