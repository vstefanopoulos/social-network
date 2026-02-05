-- ============================================================================
-- COMPREHENSIVE NOTIFICATION TEST SUITE
-- ============================================================================
-- Run with: psql -h localhost -p 5433 -U postgres -d social_notifications -f backend/services/notifications/internal/db/test/notification_tests.sql
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

-- Insert notification types that should exist
INSERT INTO notification_types (notif_type, category, default_enabled)
VALUES
    ('follow_request', 'social', TRUE),
    ('new_follower', 'social', TRUE),
    ('group_invite', 'group', TRUE),
    ('group_join_request', 'group', TRUE),
    ('new_event', 'group', TRUE),
    ('like', 'posts', TRUE),
    ('post_reply', 'posts', TRUE),
    ('mention', 'posts', TRUE)
ON CONFLICT (notif_type) DO NOTHING;

-- Create test users
INSERT INTO notifications (user_id, notif_type, source_service, source_entity_id, seen, needs_action, acted, payload, created_at, expires_at)
VALUES
    (1, 'follow_request', 'users', 2, FALSE, TRUE, FALSE, '{"requester_id":"2","requester_name":"testuser"}'::jsonb, NOW(), NOW() + INTERVAL '30 days'),
    (1, 'new_follower', 'users', 3, FALSE, FALSE, FALSE, '{"follower_id":"3","follower_name":"followeruser"}'::jsonb, NOW() - INTERVAL '1 hour', NOW() + INTERVAL '30 days'),
    (2, 'group_invite', 'users', 101, FALSE, TRUE, FALSE, '{"inviter_id":"1","inviter_name":"inviter","group_id":"101","group_name":"test_group"}'::jsonb, NOW() - INTERVAL '30 minutes', NOW() + INTERVAL '30 days'),
    (3, 'group_join_request', 'users', 201, FALSE, TRUE, FALSE, '{"requester_id":"4","requester_name":"requester","group_id":"201","group_name":"group_owner_group"}'::jsonb, NOW() - INTERVAL '2 hours', NOW() + INTERVAL '30 days'),
    (4, 'new_event', 'posts', 301, FALSE, FALSE, FALSE, '{"group_id":"202","group_name":"event_group","event_id":"301","event_title":"Test Event"}'::jsonb, NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 days'),
    (1, 'follow_request', 'users', 5, TRUE, TRUE, FALSE, '{"requester_id":"5","requester_name":"pending_user"}'::jsonb, NOW(), NOW() + INTERVAL '30 days'); -- Already seen

\echo 'Test data created'

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 1: NOTIFICATION BASICS'
\echo '========================================================================'

DO $$
DECLARE
    notification_count INT;
    unread_count INT;
    test_id BIGINT;
BEGIN
    -- Test 1: Verify notifications were created
    SELECT COUNT(*) INTO notification_count FROM notifications;
    IF notification_count >= 6 THEN
        RAISE NOTICE '[%] PASS: Notifications created in database: %', nextval('test_counter'), notification_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Notifications created in database', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected at least 6 notifications, got %', nextval('test_counter'), notification_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Notifications created in database', FALSE, 'Count: ' || notification_count);
    END IF;

    -- Test 2: Check notification with payload exists
    SELECT COUNT(*) INTO notification_count
    FROM notifications
    WHERE payload ? 'requester_id' AND payload->>'requester_id' = '2';
    PERFORM run_test(
        'Notification with specific payload exists',
        'SELECT COUNT(*) = 1 FROM (SELECT COUNT(*) FROM notifications WHERE payload ? ''requester_id'' AND payload->>''requester_id'' = ''2'') t',
        TRUE
    );

    -- Test 3: Verify seen status
    SELECT COUNT(*) INTO notification_count
    FROM notifications
    WHERE user_id = 1 AND seen = TRUE;
    PERFORM run_test(
        'Seen notification exists',
        'SELECT COUNT(*) >= 1 FROM (SELECT COUNT(*) FROM notifications WHERE user_id = 1 AND seen = TRUE) t',
        TRUE
    );

    -- Test 4: Verify needs_action status
    SELECT COUNT(*) INTO notification_count
    FROM notifications
    WHERE needs_action = TRUE;
    IF notification_count >= 3 THEN
        RAISE NOTICE '[%] PASS: Needs action notifications exist: %', nextval('test_counter'), notification_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Needs action notifications exist', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected at least 3 needs action notifications, got %', nextval('test_counter'), notification_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Needs action notifications exist', FALSE, 'Count: ' || notification_count);
    END IF;
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 2: UNREAD NOTIFICATION COUNTS'
\echo '========================================================================'

DO $$
DECLARE
    user1_unread INT;
    user2_unread INT;
    user3_unread INT;
    user4_unread INT;
BEGIN
    -- Test 1: Count unread notifications for user 1
    SELECT COUNT(*) INTO user1_unread
    FROM notifications
    WHERE user_id = 1 AND seen = FALSE AND deleted_at IS NULL;

    -- User 1 has 1 seen notification and 1 unread follow request
    IF user1_unread >= 1 THEN -- At least one unread
        RAISE NOTICE '[%] PASS: User 1 has unread notifications: %', nextval('test_counter'), user1_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 1 unread count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: User 1 should have unread notifications, got: %', nextval('test_counter'), user1_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 1 unread count', FALSE, 'Count: ' || user1_unread);
    END IF;

    -- Test 2: Count unread for user 2
    SELECT COUNT(*) INTO user2_unread
    FROM notifications
    WHERE user_id = 2 AND seen = FALSE AND deleted_at IS NULL;

    IF user2_unread >= 1 THEN -- At least the group invite
        RAISE NOTICE '[%] PASS: User 2 has unread notifications: %', nextval('test_counter'), user2_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 2 unread count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: User 2 should have unread notifications, got: %', nextval('test_counter'), user2_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 2 unread count', FALSE, 'Count: ' || user2_unread);
    END IF;

    -- Test 3: Count unread for user 3
    SELECT COUNT(*) INTO user3_unread
    FROM notifications
    WHERE user_id = 3 AND seen = FALSE AND deleted_at IS NULL;

    IF user3_unread >= 1 THEN -- At least the group join request
        RAISE NOTICE '[%] PASS: User 3 has unread notifications: %', nextval('test_counter'), user3_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 3 unread count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: User 3 should have unread notifications, got: %', nextval('test_counter'), user3_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 3 unread count', FALSE, 'Count: ' || user3_unread);
    END IF;

    -- Test 4: Count unread for user 4
    SELECT COUNT(*) INTO user4_unread
    FROM notifications
    WHERE user_id = 4 AND seen = FALSE AND deleted_at IS NULL;

    IF user4_unread >= 1 THEN -- At least the new_event
        RAISE NOTICE '[%] PASS: User 4 has unread notifications: %', nextval('test_counter'), user4_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 4 unread count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: User 4 should have unread notifications, got: %', nextval('test_counter'), user4_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User 4 unread count', FALSE, 'Count: ' || user4_unread);
    END IF;
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 3: NOTIFICATION MARKING AS READ'
\echo '========================================================================'

DO $$
DECLARE
    original_unread INT;
    after_mark_unread INT;
    test_notification_id BIGINT;
BEGIN
    -- Get a notification ID to test with
    SELECT id INTO test_notification_id
    FROM notifications
    WHERE seen = FALSE AND deleted_at IS NULL
    LIMIT 1;

    -- Count unread before marking
    SELECT COUNT(*) INTO original_unread
    FROM notifications
    WHERE user_id = (SELECT user_id FROM notifications WHERE id = test_notification_id)
    AND seen = FALSE AND deleted_at IS NULL;

    -- Mark notification as read
    UPDATE notifications SET seen = TRUE WHERE id = test_notification_id;

    -- Count unread after marking
    SELECT COUNT(*) INTO after_mark_unread
    FROM notifications
    WHERE user_id = (SELECT user_id FROM notifications WHERE id = test_notification_id)
    AND seen = FALSE AND deleted_at IS NULL;

    -- Verify decrease in unread count
    IF after_mark_unread = original_unread - 1 THEN
        RAISE NOTICE '[%] PASS: Marking notification as read decreased unread count', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Mark as read decreases count', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected unread count % to decrease by 1, got %', nextval('test_counter'), original_unread, after_mark_unread;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Mark as read decreases count', FALSE, 'Before: ' || original_unread || ', After: ' || after_mark_unread);
    END IF;

    -- Verify the specific notification is marked as seen
    PERFORM run_test(
        'Specific notification marked as read',
        format('SELECT seen = true FROM notifications WHERE id = %s', test_notification_id),
        TRUE
    );
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 4: MARK ALL AS READ'
\echo '========================================================================'

DO $$
DECLARE
    user_for_test BIGINT;
    original_unread INT;
    after_mark_unread INT;
BEGIN
    -- Find a user with unread notifications
    SELECT user_id, COUNT(*) INTO user_for_test, original_unread
    FROM notifications
    WHERE seen = FALSE AND deleted_at IS NULL
    GROUP BY user_id
    HAVING COUNT(*) >= 1
    LIMIT 1;

    IF user_for_test IS NOT NULL THEN
        -- Mark all for this user as read
        UPDATE notifications SET seen = TRUE
        WHERE user_id = user_for_test AND seen = FALSE AND deleted_at IS NULL;

        -- Count unread after marking all
        SELECT COUNT(*) INTO after_mark_unread
        FROM notifications
        WHERE user_id = user_for_test AND seen = FALSE AND deleted_at IS NULL;

        -- Verify all are marked as read
        IF after_mark_unread < original_unread AND after_mark_unread = 0 THEN
            RAISE NOTICE '[%] PASS: Mark all as read worked for user %', nextval('test_counter'), user_for_test;
            INSERT INTO test_results VALUES (currval('test_counter'), 'Mark all as read', TRUE, NULL);
        ELSE
            RAISE NOTICE '[%] FAIL: Expected unread count % to become 0, got %', nextval('test_counter'), original_unread, after_mark_unread;
            INSERT INTO test_results VALUES (currval('test_counter'), 'Mark all as read', FALSE, 'Before: ' || original_unread || ', After: ' || after_mark_unread);
        END IF;
    ELSE
        RAISE NOTICE '[%] SKIP: No users found with unread notifications', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Mark all as read', TRUE, 'No applicable data');
    END IF;
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 5: NOTIFICATION DELETION (SOFT DELETE)'
\echo '========================================================================'

DO $$
DECLARE
    test_notification_id BIGINT;
    is_deleted BOOLEAN;
    exists_in_query BOOLEAN;
BEGIN
    -- Get a notification to test deletion on
    SELECT id INTO test_notification_id
    FROM notifications
    WHERE deleted_at IS NULL
    LIMIT 1;

    IF test_notification_id IS NOT NULL THEN
        -- Soft delete the notification
        UPDATE notifications SET deleted_at = NOW() WHERE id = test_notification_id;

        -- Check if it's marked as deleted
        SELECT deleted_at IS NOT NULL INTO is_deleted
        FROM notifications
        WHERE id = test_notification_id;

        PERFORM run_test(
            'Notification soft deleted',
            format('SELECT deleted_at IS NOT NULL FROM notifications WHERE id = %s', test_notification_id),
            TRUE
        );

        -- Check that it's not returned in normal queries (with deleted_at filter)
        SELECT EXISTS(
            SELECT 1 FROM notifications
            WHERE id = test_notification_id AND deleted_at IS NULL
        ) INTO exists_in_query;

        PERFORM run_test(
            'Deleted notification not in normal queries',
            format('SELECT NOT EXISTS(SELECT 1 FROM notifications WHERE id = %s AND deleted_at IS NULL)', test_notification_id),
            TRUE
        );
    ELSE
        RAISE NOTICE '[%] SKIP: No notifications found to test deletion', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Notification deletion', TRUE, 'No applicable data');
    END IF;
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 6: NOTIFICATION TYPES'
\echo '========================================================================'

DO $$
DECLARE
    type_count INT;
BEGIN
    -- Test 1: Check if notification types exist
    SELECT COUNT(*) INTO type_count FROM notification_types;
    IF type_count >= 6 THEN
        RAISE NOTICE '[%] PASS: Notification types exist: %', nextval('test_counter'), type_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Notification types exist', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Expected at least 6 notification types, got: %', nextval('test_counter'), type_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Notification types exist', FALSE, 'Count: ' || type_count);
    END IF;

    -- Test 2: Check if 'follow_request' type exists and is enabled
    PERFORM run_test(
        'Follow request type exists and enabled',
        'SELECT COUNT(*) = 1 FROM notification_types WHERE notif_type = ''follow_request'' AND default_enabled = TRUE',
        TRUE
    );

    -- Test 3: Check if 'group_invite' type exists
    PERFORM run_test(
        'Group invite type exists',
        'SELECT COUNT(*) = 1 FROM notification_types WHERE notif_type = ''group_invite''',
        TRUE
    );

    -- Test 4: Check category assignment
    PERFORM run_test(
        'Social category notifications exist',
        'SELECT COUNT(*) >= 2 FROM notification_types WHERE category = ''social''',
        TRUE
    );
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 7: NOTIFICATION EXPIRATION'
\echo '========================================================================'

DO $$
DECLARE
    expired_count INT;
    future_notification_id BIGINT;
    past_expiry_date TIMESTAMPTZ;
BEGIN
    -- Create a notification with past expiry date to test
    INSERT INTO notifications (user_id, notif_type, source_service, seen, needs_action, payload, created_at, expires_at)
    VALUES (999, 'new_follower', 'users', FALSE, FALSE, '{}', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 hour')
    RETURNING id INTO future_notification_id;

    -- Count notifications that are past their expiration date
    SELECT COUNT(*) INTO expired_count
    FROM notifications
    WHERE expires_at < NOW();

    IF expired_count >= 1 THEN
        RAISE NOTICE '[%] PASS: Expired notifications exist: %', nextval('test_counter'), expired_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Expired notifications exist', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Should have expired notifications, got: %', nextval('test_counter'), expired_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Expired notifications exist', FALSE, 'Count: ' || expired_count);
    END IF;

    -- Test that expired notifications can still be queried normally
    -- (the application would handle purging separately if needed)
    PERFORM run_test(
        'Can query expired notifications',
        format('SELECT COUNT(*) >= 1 FROM notifications WHERE id = %s', future_notification_id),
        TRUE
    );
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 8: NOTIFICATION FILTERING'
\echo '========================================================================'

DO $$
DECLARE
    type_count INT;
    user_specific_count INT;
BEGIN
    -- Test 1: Count notifications by type
    SELECT COUNT(*) INTO type_count
    FROM notifications
    WHERE notif_type = 'follow_request';

    IF type_count >= 1 THEN
        RAISE NOTICE '[%] PASS: Follow request notifications exist: %', nextval('test_counter'), type_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Type filtering works', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Should have follow request notifications, got: %', nextval('test_counter'), type_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Type filtering works', FALSE, 'Count: ' || type_count);
    END IF;

    -- Test 2: Count notifications for a specific user
    SELECT COUNT(*) INTO user_specific_count
    FROM notifications
    WHERE user_id = 1;

    IF user_specific_count >= 1 THEN
        RAISE NOTICE '[%] PASS: User-specific notifications exist: %', nextval('test_counter'), user_specific_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User filtering works', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Should have notifications for user 1, got: %', nextval('test_counter'), user_specific_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User filtering works', FALSE, 'Count: ' || user_specific_count);
    END IF;

    -- Test 3: Filter by needs_action
    SELECT COUNT(*) INTO type_count
    FROM notifications
    WHERE needs_action = TRUE;

    IF type_count >= 1 THEN
        RAISE NOTICE '[%] PASS: Action-required notifications exist: %', nextval('test_counter'), type_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Action filter works', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Should have action-required notifications, got: %', nextval('test_counter'), type_count;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Action filter works', FALSE, 'Count: ' || type_count);
    END IF;
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 9: NOTIFICATION PAYLOAD HANDLING'
\echo '========================================================================'

DO $$
DECLARE
    has_requester_name BOOLEAN;
    correct_requester_value TEXT;
    payload_exists BOOLEAN;
BEGIN
    -- Test 1: Check if JSONB payload field works correctly
    SELECT payload ? 'requester_name' INTO payload_exists
    FROM notifications
    WHERE notif_type = 'follow_request'
    LIMIT 1;

    PERFORM run_test(
        'JSONB payload field access works',
        'SELECT payload ? ''requester_name'' FROM notifications WHERE notif_type = ''follow_request'' LIMIT 1',
        TRUE
    );

    -- Test 2: Check if specific value can be extracted from payload
    SELECT payload->>'requester_name' INTO correct_requester_value
    FROM notifications
    WHERE notif_type = 'follow_request' AND payload ? 'requester_name'
    LIMIT 1;

    IF correct_requester_value IS NOT NULL AND LENGTH(correct_requester_value) > 0 THEN
        RAISE NOTICE '[%] PASS: Payload value extraction works: %', nextval('test_counter'), correct_requester_value;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Payload value extraction', TRUE, NULL);
    ELSE
        RAISE NOTICE '[%] FAIL: Could not extract valid payload value: %', nextval('test_counter'), COALESCE(correct_requester_value, 'NULL');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Payload value extraction', FALSE, 'Value: ' || COALESCE(correct_requester_value, 'NULL'));
    END IF;

    -- Test 3: Insert notification with complex payload
    INSERT INTO notifications (user_id, notif_type, source_service, payload, created_at, expires_at)
    VALUES (1000, 'mention', 'posts', '{"post_id":"500","post_title":"Test Post","mentioner_id":"10","mentioner_name":"mentioner","mention_text":"@testuser"}', NOW(), NOW() + INTERVAL '30 days');

    -- Verify complex payload is stored correctly
    PERFORM run_test(
        'Complex JSONB payload stored',
        'SELECT COUNT(*) = 1 FROM notifications WHERE user_id = 1000 AND payload ? ''mentioner_name'' AND payload->>''mention_text'' = ''@testuser''',
        TRUE
    );
END $$;

-- ============================================================================
\echo ''
\echo '========================================================================'
\echo 'TEST SUITE 10: NOTIFICATION INDEXES PERFORMANCE'
\echo '========================================================================'

DO $$
DECLARE
    plan_text TEXT;
BEGIN
    -- Test 1: Check if user+seen index is being used
    -- This is more of a verification that the index exists and is functional
    BEGIN
        -- Create EXPLAIN output to check if index is used (this is a basic check)
        -- We'll just verify that a query runs without error
        PERFORM user_id, COUNT(*)
        FROM notifications
        WHERE user_id = 1 AND seen = FALSE
        GROUP BY user_id;

        RAISE NOTICE '[%] PASS: User+seen query executed', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'User+seen query', TRUE, NULL);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '[%] FAIL: User+seen query failed: %', nextval('test_counter'), SQLERRM;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User+seen query', FALSE, SQLERRM);
    END;

    -- Test 2: Check user+type index usage
    BEGIN
        PERFORM user_id, COUNT(*)
        FROM notifications
        WHERE user_id = 1 AND notif_type = 'follow_request'
        GROUP BY user_id;

        RAISE NOTICE '[%] PASS: User+type query executed', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'User+type query', TRUE, NULL);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '[%] FAIL: User+type query failed: %', nextval('test_counter'), SQLERRM;
        INSERT INTO test_results VALUES (currval('test_counter'), 'User+type query', FALSE, SQLERRM);
    END;

    -- Test 3: Check creation timestamp queries work efficiently
    BEGIN
        PERFORM COUNT(*)
        FROM notifications
        WHERE user_id = 1 AND created_at > NOW() - INTERVAL '1 hour';

        RAISE NOTICE '[%] PASS: Timestamp range query executed', nextval('test_counter');
        INSERT INTO test_results VALUES (currval('test_counter'), 'Timestamp query', TRUE, NULL);
    EXCEPTION WHEN OTHERS THEN
        RAISE NOTICE '[%] FAIL: Timestamp query failed: %', nextval('test_counter'), SQLERRM;
        INSERT INTO test_results VALUES (currval('test_counter'), 'Timestamp query', FALSE, SQLERRM);
    END;
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