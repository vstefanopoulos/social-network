package dbservice

const (
	// ====================================
	// GROUP_CONVERSATIONS
	// ====================================

	createGroupMessage = `
	WITH upsert_conversation AS (
		INSERT INTO group_conversations (group_id)
		VALUES ($1)
		ON CONFLICT (group_id) DO UPDATE
			SET updated_at = CURRENT_TIMESTAMP
		WHERE group_conversations.deleted_at IS NULL
		RETURNING group_id
	)
	INSERT INTO group_messages (group_id, sender_id, message_text)
	SELECT group_id, $2, $3
	FROM upsert_conversation
	RETURNING
		id,
		group_id,
		sender_id,
		message_text,
		created_at,
		updated_at,
		deleted_at;
	`

	getPrevGroupMsgs = `
    SELECT
        gm.id,
        gm.group_id,
        gm.sender_id,
        gm.message_text,
        gm.created_at,
        gm.updated_at,
        gm.deleted_at

    FROM group_conversations gc
    JOIN group_messages gm
        ON gm.group_id = gc.group_id
    WHERE gc.group_id = $1
      AND gm.deleted_at IS NULL
	  AND gc.deleted_at IS NULL
      AND gm.id < $2
    ORDER BY gm.id DESC
    LIMIT $3;
	`

	getNextGroupMsgs = `
    SELECT
        gm.id,
        gm.group_id,
        gm.sender_id,
        gm.message_text,
        gm.created_at,
        gm.updated_at,
        gm.deleted_at

    FROM group_conversations gc
    JOIN group_messages gm
        ON gm.group_id = gc.group_id
    WHERE gc.group_id = $1
      AND gm.deleted_at IS NULL
	  AND gc.deleted_at IS NULL
      AND gm.id > $2
    ORDER BY gm.id ASC
    LIMIT $3;
	`

	// ====================================
	// PRIVATE_CONVERSATIONS
	// ====================================

	createMsgAndConv = `
	WITH conv AS (
		INSERT INTO private_conversations (user_a, user_b)
		VALUES ($1, $2)
		ON CONFLICT (user_a, user_b)
		DO UPDATE SET user_a = private_conversations.user_a
		RETURNING id, user_a, user_b, deleted_at
	),
	inserted_message AS (
		INSERT INTO private_messages (conversation_id, sender_id, message_text)
		SELECT c.id, $3, $4
		FROM conv c
		WHERE c.deleted_at IS NULL
		RETURNING *
	)
	SELECT 
		im.id,
		im.conversation_id,
		im.sender_id,
		im.message_text,
		im.created_at,
		im.updated_at,
		im.deleted_at
	FROM inserted_message im;
	`

	getConvsWithUnreadsCount = `
	SELECT COUNT(*) AS unread_conversations_count
	FROM private_conversations pc
	JOIN LATERAL (
		SELECT pm.id
		FROM private_messages pm
		WHERE pm.conversation_id = pc.id
		ORDER BY pm.id DESC
		LIMIT 1
	) latest_message ON true
	WHERE pc.deleted_at IS NULL
	AND (
		(pc.user_a = $1 AND (pc.last_read_message_id_a IS NULL OR latest_message.id > pc.last_read_message_id_a))
		OR
		(pc.user_b = $1 AND (pc.last_read_message_id_b IS NULL OR latest_message.id > pc.last_read_message_id_b))
	);
	`

	getPrivateConvById = `
	WITH user_conversation AS (
		SELECT
			pc.id AS conversation_id,
			pc.updated_at,

			CASE
				WHEN pc.user_a = $1 THEN pc.user_b
				ELSE pc.user_a
			END AS other_user_id,

			CASE
				WHEN pc.user_a = $1 THEN pc.last_read_message_id_a
				ELSE pc.last_read_message_id_b
			END AS last_read_message_id

		FROM private_conversations pc
		WHERE pc.id = $2
			AND $1 IN (pc.user_a, pc.user_b)
	)

	SELECT
		uc.conversation_id,
		uc.updated_at,
		uc.other_user_id,

		lm.id           AS last_message_id,
		lm.sender_id    AS last_message_sender_id,
		lm.message_text AS last_message_text,
		lm.created_at   AS last_message_created_at,

		COUNT(pm.id) FILTER (
			WHERE pm.id > COALESCE(uc.last_read_message_id, 0)
		) AS unread_count

	FROM user_conversation uc

	LEFT JOIN LATERAL (
		SELECT pm.id, pm.sender_id, pm.message_text, pm.created_at
		FROM private_messages pm
		WHERE pm.conversation_id = uc.conversation_id
			AND pm.deleted_at IS NULL
		ORDER BY pm.id DESC
		LIMIT 1
	) lm ON true

	LEFT JOIN private_messages pm
		ON pm.conversation_id = uc.conversation_id
		AND pm.deleted_at IS NULL

	GROUP BY
		uc.conversation_id,
		uc.updated_at,
		uc.other_user_id,
		uc.last_read_message_id,
		lm.id,
		lm.sender_id,
		lm.message_text,
		lm.created_at;
	`

	getPrivateConvs = `
	WITH user_conversations AS (
		SELECT
			pc.id AS conversation_id,
			pc.updated_at,

			-- determine other user
			CASE
				WHEN pc.user_a = $1::bigint THEN pc.user_b
				ELSE pc.user_a
			END AS other_user_id,

			-- determine last read message for this user
			CASE
				WHEN pc.user_a = $1::bigint THEN pc.last_read_message_id_a
				ELSE pc.last_read_message_id_b
			END AS last_read_message_id

		FROM private_conversations pc
		WHERE $1 IN (pc.user_a, pc.user_b)
		AND pc.updated_at < $2
	)

	SELECT
		uc.conversation_id,
		uc.updated_at,
		uc.other_user_id,

		-- last message
		lm.id           AS last_message_id,
		lm.sender_id    AS last_message_sender_id,
		lm.message_text AS last_message_text,
		lm.created_at   AS last_message_created_at,

		-- unread count
		COUNT(pm.id) FILTER (
			WHERE pm.id > COALESCE(uc.last_read_message_id, 0)
		) AS unread_count

	FROM user_conversations uc

	-- last message per conversation
	LEFT JOIN LATERAL (
		SELECT pm.id, pm.sender_id, pm.message_text, pm.created_at
		FROM private_messages pm
		WHERE pm.conversation_id = uc.conversation_id
		AND pm.deleted_at IS NULL
		ORDER BY pm.id DESC
		LIMIT 1
	) lm ON true

	-- unread messages
	LEFT JOIN private_messages pm
		ON pm.conversation_id = uc.conversation_id
	AND pm.deleted_at IS NULL

	GROUP BY
		uc.conversation_id,
		uc.updated_at,
		uc.other_user_id,
		uc.last_read_message_id,
		lm.id,
		lm.sender_id,
		lm.message_text,
		lm.created_at

	ORDER BY uc.updated_at DESC
	LIMIT $3;
	`

	getPrevPrivateMsgs = `
	SELECT pm.*
	FROM private_conversations pc
	JOIN private_messages pm
	ON pm.conversation_id = pc.id
	WHERE pc.deleted_at IS NULL
	AND pm.deleted_at IS NULL
	AND pm.id < $3
	AND (
		(pc.user_a = LEAST($1::bigint, $2::bigint) AND pc.user_b = GREATEST($1::bigint, $2::bigint))
		)
	ORDER BY pm.id DESC
	LIMIT $4;
	`

	getNextPrivateMsgs = `
	SELECT pm.*
	FROM private_conversations pc
	JOIN private_messages pm
	ON pm.conversation_id = pc.id
	WHERE pc.deleted_at IS NULL
	AND pm.deleted_at IS NULL
	AND pm.id > $3
	AND (
		(pc.user_a = LEAST($1::bigint, $2::bigint) AND pc.user_b = GREATEST($1::bigint, $2::bigint))
		)
	ORDER BY pm.id ASC
	LIMIT $4;
	`

	updateLastReadMessage = `
	UPDATE private_conversations
	SET
		last_read_message_id_a = CASE
			WHEN user_a = $2 THEN $3
			ELSE last_read_message_id_a
		END,
		last_read_message_id_b = CASE
			WHEN user_b = $2 THEN $3
			ELSE last_read_message_id_b
		END
	WHERE id = $1
	AND (user_a = $2 OR user_b = $2)
	AND deleted_at IS NULL;
	`
)
