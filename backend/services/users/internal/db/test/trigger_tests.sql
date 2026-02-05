-- ============================================================================
-- COMPREHENSIVE TRIGGER TEST SUITE
-- ============================================================================
-- Run with: psql -h localhost -p 5433 -U postgres -d social_users -f backend/services/users/internal/db/test/trigger_tests.sql
-- ============================================================================

\set ON_ERROR_STOP on
\timing off

BEGIN;

-- Test result tracking
CREATE TEMP TABLE test_results (
    test_number INT,
    test_name TEXT,
    passed BOOLEAN,
    error_message TEXT
);

CREATE TEMP SEQUENCE test_counter;

-- Helper function to run tests
CREATE OR REPLACE FUNCTION run_test(
    test_name TEXT,
    test_query TEXT,
    expected_result BOOLEAN,
    error_msg TEXT DEFAULT NULL
) RETURNS VOID AS $$
DECLARE
    test_num INT;
    actual_result BOOLEAN;
BEGIN
    test_num := nextval('test_counter');
    
    BEGIN
        EXECUTE test_query INTO actual_result;
        
        IF actual_result = expected_result THEN
            INSERT INTO test_results VALUES (test_num, test_name, TRUE, NULL);
            RAISE NOTICE '[%] PASS: %', test_num, test_name;
        ELSE
            INSERT INTO test_results VALUES (test_num, test_name, FALSE, 
                format('Expected %s but got %s', expected_result, actual_result));
            RAISE NOTICE '[%] FAIL: % - Expected % but got %', 
                test_num, test_name, expected_result, actual_result;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        INSERT INTO test_results VALUES (test_num, test_name, FALSE, SQLERRM);
        RAISE NOTICE '[%] FAIL: % - Error: %', test_num, test_name, SQLERRM;
    END;
END;
$$ LANGUAGE plpgsql;

-- Helper to expect exception
CREATE OR REPLACE FUNCTION expect_exception(
    test_name TEXT,
    test_query TEXT,
    expected_error_substring TEXT DEFAULT NULL
) RETURNS VOID AS $$
DECLARE
    test_num INT;
BEGIN
    test_num := nextval('test_counter');
    
    BEGIN
        EXECUTE test_query;
        -- If we get here, no exception was raised
        INSERT INTO test_results VALUES (test_num, test_name, FALSE, 
            'Expected exception but none was raised');
        RAISE NOTICE '[%] FAIL: % - Expected exception but none was raised', 
            test_num, test_name;
    EXCEPTION WHEN OTHERS THEN
        IF expected_error_substring IS NULL OR SQLERRM LIKE '%' || expected_error_substring || '%' THEN
            INSERT INTO test_results VALUES (test_num, test_name, TRUE, NULL);
            RAISE NOTICE '[%] PASS: % (caught expected exception)', test_num, test_name;
        ELSE
            INSERT INTO test_results VALUES (test_num, test_name, FALSE, 
                format('Expected error containing "%s" but got: %s', expected_error_substring, SQLERRM));
            RAISE NOTICE '[%] FAIL: % - Wrong error message: %', test_num, test_name, SQLERRM;
        END IF;
    END;
END;
$$ LANGUAGE plpgsql;

\echo ''
\echo '========================================================================'
\echo 'SETTING UP TEST DATA'
\echo '========================================================================'

-- Create test users
INSERT INTO users (username, first_name, last_name, date_of_birth, avatar, about_me, profile_public, current_status)
VALUES 
    ('alice', 'Alice', 'Test', '1990-01-01', 'alice.jpg', 'Alice bio', TRUE, 'active'),
    ('bob', 'Bob', 'Test', '1991-01-01', 'bob.jpg', 'Bob bio', FALSE, 'active'),
    ('charlie', 'Charlie', 'Test', '1992-01-01', 'charlie.jpg', 'Charlie bio', TRUE, 'active'),
    ('diana', 'Diana', 'Test', '1993-01-01', 'diana.jpg', 'Diana bio', FALSE, 'active');

\echo 'Test users created'

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 1: UPDATED_AT TRIGGERS'
\echo '========================================================================'

DO $$
DECLARE
    user_id BIGINT;
    old_timestamp TIMESTAMPTZ;
    new_timestamp TIMESTAMPTZ;
BEGIN
    -- Get a user
    SELECT id INTO user_id FROM users WHERE username = 'alice';
    
    -- Record current updated_at (should be NULL initially)
    SELECT updated_at INTO old_timestamp FROM users WHERE id = user_id;
    
    -- Wait a tiny bit to ensure timestamp difference
    PERFORM pg_sleep(0.01);
    
    -- Update the user
    UPDATE users SET first_name = 'Alice Updated' WHERE id = user_id;
    
    -- Check updated_at was set
    SELECT updated_at INTO new_timestamp FROM users WHERE id = user_id;
    
    IF new_timestamp IS NOT NULL AND new_timestamp > COALESCE(old_timestamp, TIMESTAMP '2000-01-01') THEN
        RAISE NOTICE '[%] PASS: updated_at trigger sets timestamp on update', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'updated_at trigger sets timestamp', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: updated_at not set properly', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'updated_at trigger sets timestamp', FALSE, 'Timestamp not updated');
    END IF;
    
    -- Test that updated_at doesn't change if data doesn't change
    old_timestamp := new_timestamp;
    PERFORM pg_sleep(0.01);
    UPDATE users SET first_name = 'Alice Updated' WHERE id = user_id; -- Same value
    SELECT updated_at INTO new_timestamp FROM users WHERE id = user_id;
    
    IF new_timestamp = old_timestamp THEN
        RAISE NOTICE '[%] PASS: updated_at not changed when data unchanged', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'updated_at unchanged for identical update', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: updated_at changed even though data unchanged', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'updated_at unchanged for identical update', FALSE, 'Timestamp changed unnecessarily');
    END IF;
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 2: FOLLOW SYSTEM'
\echo '========================================================================'

DO $$
DECLARE
    alice_id BIGINT;
    bob_id BIGINT;
    charlie_id BIGINT;
    result TEXT;
    follow_count INT;
    request_count INT;
BEGIN
    SELECT id INTO alice_id FROM users WHERE username = 'alice'; -- public
    SELECT id INTO bob_id FROM users WHERE username = 'bob';     -- private
    SELECT id INTO charlie_id FROM users WHERE username = 'charlie'; -- public
    
    -- Test 1: Follow public user succeeds
    result := follow_user(alice_id, charlie_id);
    IF result = 'followed' THEN
        RAISE NOTICE '[%] PASS: Can follow public user', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow public user', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Follow public user returned: %', nextval('test_counter'), result;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow public user', FALSE, 'Got: ' || result);
    END IF;
    
    -- Test 2: Follow relationship exists
    SELECT COUNT(*) INTO follow_count FROM follows WHERE follower_id = alice_id AND following_id = charlie_id;
    IF follow_count = 1 THEN
        RAISE NOTICE '[%] PASS: Follow exists in database', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow exists in DB', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Follow not in database', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow exists in DB', FALSE, 'Count: ' || follow_count);
    END IF;
    
    -- Test 3: Already following returns correct message
    result := follow_user(alice_id, charlie_id);
    IF result = 'already_following' THEN
        RAISE NOTICE '[%] PASS: Already following detected', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Already following detection', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Should return already_following, got: %', nextval('test_counter'), result;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Already following detection', FALSE, 'Got: ' || result);
    END IF;
    
    -- Test 4: Follow private user creates request
    result := follow_user(alice_id, bob_id);
    IF result = 'requested' THEN
        RAISE NOTICE '[%] PASS: Private user creates follow request', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow private user creates request', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Should return requested, got: %', nextval('test_counter'), result;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow private user creates request', FALSE, 'Got: ' || result);
    END IF;
    
    -- Test 5: Follow request exists with pending status
    SELECT COUNT(*) INTO request_count 
    FROM follow_requests 
    WHERE requester_id = alice_id AND target_id = bob_id AND status = 'pending';
    IF request_count = 1 THEN
        RAISE NOTICE '[%] PASS: Follow request exists with pending status', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow request pending', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Follow request not found or wrong status', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow request pending', FALSE, 'Count: ' || request_count);
    END IF;
    
    -- Test 6: Duplicate request returns already pending
    result := follow_user(alice_id, bob_id);
    IF result = 'request_already_pending' THEN
        RAISE NOTICE '[%] PASS: Duplicate request detected', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Duplicate request detection', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Should return request_already_pending, got: %', nextval('test_counter'), result;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Duplicate request detection', FALSE, 'Got: ' || result);
    END IF;
    
    -- Test 7: Accepting request creates follow
    UPDATE follow_requests 
    SET status = 'accepted' 
    WHERE requester_id = alice_id AND target_id = bob_id;
    
    SELECT COUNT(*) INTO follow_count 
    FROM follows 
    WHERE follower_id = alice_id AND following_id = bob_id;
    
    IF follow_count = 1 THEN
        RAISE NOTICE '[%] PASS: Accepting request creates follow', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Accept request creates follow', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Follow not created after accepting request', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Accept request creates follow', FALSE, 'Count: ' || follow_count);
    END IF;
    
    -- Test 8: Profile going public accepts pending requests
    -- Create new request from charlie to bob
    result := follow_user(charlie_id, bob_id);
    
    -- Make bob's profile public
    UPDATE users SET profile_public = TRUE WHERE id = bob_id;
    
    -- Check request was accepted
    SELECT status INTO result 
    FROM follow_requests 
    WHERE requester_id = charlie_id AND target_id = bob_id;
    
    IF result = 'accepted' THEN
        RAISE NOTICE '[%] PASS: Profile going public accepts pending requests', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Public profile accepts requests', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Request not accepted when profile went public', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Public profile accepts requests', FALSE, 'Status: ' || result);
    END IF;
    
    -- Check follow was created
    SELECT COUNT(*) INTO follow_count 
    FROM follows 
    WHERE follower_id = charlie_id AND following_id = bob_id;
    
    IF follow_count = 1 THEN
        RAISE NOTICE '[%] PASS: Follow created when profile went public', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow created on public switch', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Follow not created when profile went public', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Follow created on public switch', FALSE, 'Count: ' || follow_count);
    END IF;
END $$;

-- Test 9: Cannot follow yourself
DO $$
BEGIN
    PERFORM expect_exception(
        'Cannot follow yourself',
        format('SELECT follow_user(%s, %s)', 
            (SELECT id FROM users WHERE username = 'alice'),
            (SELECT id FROM users WHERE username = 'alice')
        ),
        'Cannot follow yourself'
    );
END $$;

-- Test 10: Cannot follow non-existent user
DO $$
BEGIN
    PERFORM expect_exception(
        'Cannot follow non-existent user',
        'SELECT follow_user(1, 999999)',
        'not found'
    );
END $$;

-- Test 10: Cannot follow non-existent user
DO $$
BEGIN
PERFORM expect_exception(
    'Cannot follow non-existent user',
    'SELECT follow_user(1, 999999)',
    'not found'
);
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 3: GROUP SYSTEM'
\echo '========================================================================'

DO $$
DECLARE
    alice_id BIGINT;
    bob_id BIGINT;
    charlie_id BIGINT;
    test_group_id BIGINT;
    member_count INT;
    owner_exists BOOLEAN;
BEGIN
    SELECT id INTO alice_id FROM users WHERE username = 'alice';
    SELECT id INTO bob_id FROM users WHERE username = 'bob';
    SELECT id INTO charlie_id FROM users WHERE username = 'charlie';
    
    -- Test 1: Creating group initializes member count to 0
    INSERT INTO groups (group_owner, group_title, group_description)
    VALUES (alice_id, 'Test Group 1', 'Description')
    RETURNING id INTO test_group_id;
    
    -- Immediately after insert, before trigger fires
    SELECT members_count INTO member_count FROM groups WHERE id = test_group_id;
    
    -- The trigger should have fired by now and added the owner
    -- Wait a tiny bit to ensure trigger completes
    PERFORM pg_sleep(0.01);
    
    SELECT members_count INTO member_count FROM groups WHERE id = test_group_id;
    
    IF member_count = 1 THEN
        RAISE NOTICE '[%] PASS: Group member count is 1 after creation', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Group initial member count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected member_count=1, got %', nextval('test_counter'), member_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Group initial member count', FALSE, 'Count: ' || member_count);
    END IF;
    
    -- Test 2: Owner is added to group_members
    SELECT EXISTS(
        SELECT 1 FROM group_members 
        WHERE group_id = test_group_id AND user_id = alice_id AND role = 'owner'
    ) INTO owner_exists;
    
    IF owner_exists THEN
        RAISE NOTICE '[%] PASS: Owner added to group_members', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Owner in group_members', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Owner not in group_members', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Owner in group_members', FALSE, 'Owner not found');
    END IF;
    
    -- Test 3: Adding member increments count
    INSERT INTO group_members (group_id, user_id, role)
    VALUES (test_group_id, bob_id, 'member');
    
    SELECT members_count INTO member_count FROM groups WHERE id = test_group_id;
    
    IF member_count = 2 THEN
        RAISE NOTICE '[%] PASS: Adding member increments count', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Add member increments count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected member_count=2, got %', nextval('test_counter'), member_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Add member increments count', FALSE, 'Count: ' || member_count);
    END IF;
    
    -- Test 4: Soft-deleting member decrements count
    UPDATE group_members 
    SET deleted_at = CURRENT_TIMESTAMP 
    WHERE group_id = test_group_id AND user_id = bob_id;
    
    SELECT members_count INTO member_count FROM groups WHERE id = test_group_id;
    
    IF member_count = 1 THEN
        RAISE NOTICE '[%] PASS: Soft-delete decrements count', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Soft-delete decrements count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected member_count=1 after soft-delete, got %', nextval('test_counter'), member_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Soft-delete decrements count', FALSE, 'Count: ' || member_count);
    END IF;
    
    -- Test 5: Restoring member increments count
    UPDATE group_members 
    SET deleted_at = NULL 
    WHERE group_id = test_group_id AND user_id = bob_id;
    
    SELECT members_count INTO member_count FROM groups WHERE id = test_group_id;
    
    IF member_count = 2 THEN
        RAISE NOTICE '[%] PASS: Restore increments count', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Restore increments count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected member_count=2 after restore, got %', nextval('test_counter'), member_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Restore increments count', FALSE, 'Count: ' || member_count);
    END IF;
    
    -- Test 6: Join request acceptance adds member
    INSERT INTO group_join_requests (group_id, user_id, status)
    VALUES (test_group_id, charlie_id, 'pending');
    
    UPDATE group_join_requests 
    SET status = 'accepted' 
    WHERE group_id = test_group_id AND user_id = charlie_id;
    
    SELECT EXISTS(
        SELECT 1 FROM group_members 
        WHERE group_id = test_group_id AND user_id = charlie_id
    ) INTO owner_exists;
    
    IF owner_exists THEN
        RAISE NOTICE '[%] PASS: Join request acceptance adds member', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Join request adds member', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Member not added after join request accepted', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Join request adds member', FALSE, 'Member not found');
    END IF;
    
    -- Check count updated
    SELECT members_count INTO member_count FROM groups WHERE id = test_group_id;
    IF member_count = 3 THEN
        RAISE NOTICE '[%] PASS: Join request updates count', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Join request updates count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected member_count=3, got %', nextval('test_counter'), member_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Join request updates count', FALSE, 'Count: ' || member_count);
    END IF;
    
    -- Test 7: Group invite acceptance adds member
    -- Create new group and new user for clean test
    INSERT INTO groups (group_owner, group_title, group_description)
    VALUES (alice_id, 'Test Group 2', 'Description')
    RETURNING id INTO test_group_id;
    
    PERFORM pg_sleep(0.01); -- Let owner be added
    
    INSERT INTO group_invites (group_id, sender_id, receiver_id, status)
    VALUES (test_group_id, alice_id, bob_id, 'pending');
    
    UPDATE group_invites 
    SET status = 'accepted' 
    WHERE group_id = test_group_id AND receiver_id = bob_id;
    
    SELECT EXISTS(
        SELECT 1 FROM group_members 
        WHERE group_id = test_group_id AND user_id = bob_id
    ) INTO owner_exists;
    
    IF owner_exists THEN
        RAISE NOTICE '[%] PASS: Invite acceptance adds member', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Invite acceptance adds member', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Member not added after invite accepted', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Invite acceptance adds member', FALSE, 'Member not found');
    END IF;
END $$;

-- Test 8: Cannot remove group owner
DO $$
DECLARE
    alice_id BIGINT;
    test_group_id BIGINT;
BEGIN
    SELECT id INTO alice_id FROM users WHERE username = 'alice';
    SELECT id INTO test_group_id FROM groups WHERE group_title = 'Test Group 1';
    
    BEGIN
        UPDATE group_members 
        SET deleted_at = CURRENT_TIMESTAMP 
        WHERE group_id = test_group_id AND user_id = alice_id AND role = 'owner';
        
        RAISE NOTICE '[%] FAIL: Should prevent owner from leaving', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Prevent owner leave', FALSE, 'No exception raised');
    EXCEPTION WHEN OTHERS THEN
        IF SQLERRM LIKE '%owner cannot leave%' OR SQLERRM LIKE '%Transfer ownership%' THEN
            RAISE NOTICE '[%] PASS: Owner prevented from leaving', nextval('test_counter');
            INSERT INTO test_results VALUES (currval('test_counter'), 'Prevent owner leave', TRUE, NULL);
        ELSE
            RAISE NOTICE '[%] FAIL: Wrong exception: %', nextval('test_counter'), SQLERRM;
            INSERT INTO test_results VALUES (currval('test_counter'), 'Prevent owner leave', FALSE, SQLERRM);
        END IF;
    END;
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 4: SOFT DELETE CASCADE'
\echo '========================================================================'
DO $$
DECLARE
    diana_id BIGINT;
BEGIN
    -- Get Diana's user ID
    SELECT id INTO diana_id FROM users WHERE username = 'diana';

    -- Setup: Diana follows someone and has a group
    INSERT INTO follows (follower_id, following_id)
    VALUES (diana_id, (SELECT id FROM users WHERE username = 'alice'));

    INSERT INTO follow_requests (requester_id, target_id, status)
    VALUES (diana_id, (SELECT id FROM users WHERE username = 'bob'), 'pending');

    INSERT INTO groups (group_owner, group_title, group_description)
    VALUES (diana_id, 'Diana Group', 'Description');

    PERFORM pg_sleep(0.01);

    -- Attempt to soft-delete Diana; expect failure
    PERFORM expect_exception(
        'Cannot soft-delete group owner',
        format('UPDATE users SET deleted_at = CURRENT_TIMESTAMP WHERE id = %s', diana_id),
        'Group owner cannot be deleted'
    );
END $$;
-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST RESULTS SUMMARY'
\echo '========================================================================'

DO $$
DECLARE
    total_tests INT;
    passed_tests INT;
    failed_tests INT;
BEGIN
    SELECT COUNT(*) INTO total_tests FROM test_results;
    SELECT COUNT(*) INTO passed_tests FROM test_results WHERE passed = TRUE;
    SELECT COUNT(*) INTO failed_tests FROM test_results WHERE passed = FALSE;
    
    RAISE NOTICE '';
    RAISE NOTICE '========================================';
    RAISE NOTICE 'Total Tests: %', total_tests;
    RAISE NOTICE 'Passed:      % (%.1f%%)', passed_tests, (passed_tests::DECIMAL / NULLIF(total_tests, 0) * 100);
    RAISE NOTICE 'Failed:      % (%.1f%%)', failed_tests, (failed_tests::DECIMAL / NULLIF(total_tests, 0) * 100);
    RAISE NOTICE '========================================';
    RAISE NOTICE '';
    
    IF failed_tests > 0 THEN
        RAISE NOTICE 'FAILED TESTS:';
        RAISE NOTICE '';
    END IF;
END $$;

-- Show failed tests
SELECT 
    test_number,
    test_name,
    error_message
FROM test_results 
WHERE passed = FALSE 
ORDER BY test_number;

-- Final result
DO $$
DECLARE
    failed_count INT;
BEGIN
    SELECT COUNT(*) INTO failed_count FROM test_results WHERE passed = FALSE;
    
    IF failed_count = 0 THEN
        RAISE NOTICE '';
        RAISE NOTICE '✓ ALL TESTS PASSED! ✓';
        RAISE NOTICE '';
    ELSE
        RAISE NOTICE '';
        RAISE NOTICE '✗ SOME TESTS FAILED ✗';
        RAISE NOTICE '';
    END IF;
END $$;

ROLLBACK;

\echo ''
\echo 'Test completed. All changes rolled back.'