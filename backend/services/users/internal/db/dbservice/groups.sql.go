package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const acceptGroupInvite = `-- name: AcceptGroupInvite :exec
UPDATE group_invites
SET status = 'accepted'
WHERE group_id = $1
  AND receiver_id = $2
`

type AcceptGroupInviteParams struct {
	GroupID    int64
	ReceiverID int64
}

func (q *Queries) AcceptGroupInvite(ctx context.Context, arg AcceptGroupInviteParams) error {
	_, err := q.db.Exec(ctx, acceptGroupInvite, arg.GroupID, arg.ReceiverID)
	return err
}

const acceptGroupJoinRequest = `-- name: AcceptGroupJoinRequest :exec
UPDATE group_join_requests
SET status = 'accepted'
WHERE group_id = $1
  AND user_id = $2
`

type AcceptGroupJoinRequestParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) AcceptGroupJoinRequest(ctx context.Context, arg AcceptGroupJoinRequestParams) error {
	_, err := q.db.Exec(ctx, acceptGroupJoinRequest, arg.GroupID, arg.UserID)
	return err
}

const addGroupOwnerAsMember = `-- name: AddGroupOwnerAsMember :exec
INSERT INTO group_members (group_id, user_id, role)
VALUES ($1, $2, 'owner')
`

type AddGroupOwnerAsMemberParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) AddGroupOwnerAsMember(ctx context.Context, arg AddGroupOwnerAsMemberParams) error {
	_, err := q.db.Exec(ctx, addGroupOwnerAsMember, arg.GroupID, arg.UserID)
	return err
}

const addUserToGroup = `-- name: AddUserToGroup :exec
INSERT INTO group_members (group_id, user_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`

type AddUserToGroupParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) AddUserToGroup(ctx context.Context, arg AddUserToGroupParams) error {
	_, err := q.db.Exec(ctx, addUserToGroup, arg.GroupID, arg.UserID)
	return err
}

const cancelGroupInvite = `-- name: CancelGroupInvite :exec
DELETE FROM group_invites
WHERE group_id = $1
  AND receiver_id = $2
  AND sender_id=$3
`

type CancelGroupInviteParams struct {
	GroupID    int64
	ReceiverID int64
	SenderID   int64
}

func (q *Queries) CancelGroupInvite(ctx context.Context, arg CancelGroupInviteParams) error {
	_, err := q.db.Exec(ctx, cancelGroupInvite, arg.GroupID, arg.ReceiverID, arg.SenderID)
	return err
}

const cancelGroupJoinRequest = `-- name: CancelGroupJoinRequest :exec
DELETE FROM group_join_requests
WHERE group_id = $1
  AND user_id = $2
`

type CancelGroupJoinRequestParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) CancelGroupJoinRequest(ctx context.Context, arg CancelGroupJoinRequestParams) error {
	_, err := q.db.Exec(ctx, cancelGroupJoinRequest, arg.GroupID, arg.UserID)
	return err
}

const createGroup = `-- name: CreateGroup :one
INSERT INTO groups (group_owner, group_title, group_description, group_image_id)
VALUES ($1, $2, $3, $4)
RETURNING id
`

type CreateGroupParams struct {
	GroupOwner       int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
}

func (q *Queries) CreateGroup(ctx context.Context, arg CreateGroupParams) (int64, error) {
	row := q.db.QueryRow(ctx, createGroup,
		arg.GroupOwner,
		arg.GroupTitle,
		arg.GroupDescription,
		arg.GroupImageID,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const declineGroupInvite = `-- name: DeclineGroupInvite :exec
UPDATE group_invites
SET status = 'declined'
WHERE group_id = $1
  AND receiver_id = $2
`

type DeclineGroupInviteParams struct {
	GroupID    int64
	ReceiverID int64
}

func (q *Queries) DeclineGroupInvite(ctx context.Context, arg DeclineGroupInviteParams) error {
	_, err := q.db.Exec(ctx, declineGroupInvite, arg.GroupID, arg.ReceiverID)
	return err
}

const getAllGroups = `-- name: GetAllGroups :many
SELECT
  id,
  group_owner,
  group_title,
  group_description,
  group_image_id,
  members_count
FROM groups
WHERE deleted_at IS NULL
ORDER BY members_count DESC, id ASC
LIMIT $1 OFFSET $2
`

type GetAllGroupsParams struct {
	Limit  int32
	Offset int32
}

type GetAllGroupsRow struct {
	ID               int64
	GroupOwner       int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
	MembersCount     int32
}

func (q *Queries) GetAllGroups(ctx context.Context, arg GetAllGroupsParams) ([]GetAllGroupsRow, error) {
	rows, err := q.db.Query(ctx, getAllGroups, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllGroupsRow{}
	for rows.Next() {
		var i GetAllGroupsRow
		if err := rows.Scan(
			&i.ID,
			&i.GroupOwner,
			&i.GroupTitle,
			&i.GroupDescription,
			&i.GroupImageID,
			&i.MembersCount,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getGroupInfo = `-- name: GetGroupInfo :one
SELECT
  id,
  group_owner,
  group_title,
  group_description,
  group_image_id,
  members_count
FROM groups
WHERE id=$1
  AND deleted_at IS NULL
`

type GetGroupInfoRow struct {
	ID               int64
	GroupOwner       int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
	MembersCount     int32
}

func (q *Queries) GetGroupInfo(ctx context.Context, id int64) (GetGroupInfoRow, error) {
	row := q.db.QueryRow(ctx, getGroupInfo, id)
	var i GetGroupInfoRow
	err := row.Scan(
		&i.ID,
		&i.GroupOwner,
		&i.GroupTitle,
		&i.GroupDescription,
		&i.GroupImageID,
		&i.MembersCount,
	)
	return i, err
}

const getGroupBasicInfo = `-- name: GetGroupBasicInfo :one
SELECT
  id,
  group_owner,
  group_title,
  group_description,
  group_image_id
FROM groups
WHERE id=$1
  AND deleted_at IS NULL
`

type GetGroupBasicInfoRow struct {
	ID               int64
	GroupOwner       int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
}

func (q *Queries) GetGroupBasicInfo(ctx context.Context, id int64) (GetGroupBasicInfoRow, error) {
	row := q.db.QueryRow(ctx, getGroupBasicInfo, id)
	var i GetGroupBasicInfoRow
	err := row.Scan(
		&i.ID,
		&i.GroupOwner,
		&i.GroupTitle,
		&i.GroupDescription,
		&i.GroupImageID,
	)
	return i, err
}

const getGroupMembers = `-- name: GetGroupMembers :many
SELECT
    u.id,
    u.username,
    u.avatar_id,
    gm.role,
    gm.joined_at
FROM group_members gm
JOIN users u
    ON gm.user_id = u.id
WHERE gm.group_id = $1
  AND gm.deleted_at IS NULL
  AND u.deleted_at IS NULL
ORDER BY gm.joined_at DESC, u.id DESC
LIMIT $2 OFFSET $3
`

type GetGroupMembersParams struct {
	GroupID int64
	Limit   int32
	Offset  int32
}

type GetGroupMembersRow struct {
	ID       int64
	Username string
	AvatarID int64
	Role     NullGroupRole
	JoinedAt pgtype.Timestamptz
}

func (q *Queries) GetGroupMembers(ctx context.Context, arg GetGroupMembersParams) ([]GetGroupMembersRow, error) {
	rows, err := q.db.Query(ctx, getGroupMembers, arg.GroupID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetGroupMembersRow{}
	for rows.Next() {
		var i GetGroupMembersRow
		if err := rows.Scan(
			&i.ID,
			&i.Username,
			&i.AvatarID,
			&i.Role,
			&i.JoinedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getAllGroupMemberIds = `-- name: GetAllGroupMembers :many
SELECT
    u.id
FROM group_members gm
JOIN users u
    ON gm.user_id = u.id
WHERE gm.group_id = $1
  AND gm.deleted_at IS NULL
  AND u.deleted_at IS NULL
ORDER BY u.id DESC;
`

type GetAllGroupMemberIdsParams struct {
	GroupID int64
}

type GetAllGroupMemberIdsRow struct {
	ID int64
}

func (q *Queries) GetAllGroupMemberIds(ctx context.Context, arg GetAllGroupMemberIdsParams) ([]GetAllGroupMemberIdsRow, error) {
	rows, err := q.db.Query(ctx, getAllGroupMemberIds, arg.GroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetAllGroupMemberIdsRow{}
	for rows.Next() {
		var i GetAllGroupMemberIdsRow
		if err := rows.Scan(
			&i.ID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserGroupRole = `-- name: GetUserGroupRole :one
SELECT role
FROM group_members
WHERE group_id = $1
  AND user_id = $2
  AND deleted_at IS NULL
`

type GetUserGroupRoleParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) GetUserGroupRole(ctx context.Context, arg GetUserGroupRoleParams) (NullGroupRole, error) {
	row := q.db.QueryRow(ctx, getUserGroupRole, arg.GroupID, arg.UserID)
	var role NullGroupRole
	err := row.Scan(&role)
	return role, err
}

const getUserGroups = `-- name: GetUserGroups :many
SELECT
    group_id,
    group_owner,
    group_title,
    group_description,
    group_image_id,
    members_count,
    is_member,
    is_owner
FROM (
    SELECT DISTINCT
        g.id AS group_id,
        g.group_owner,
        g.group_title,
        g.group_description,
        g.group_image_id,
        g.members_count,
        CASE WHEN gm.user_id IS NOT NULL THEN TRUE ELSE FALSE END AS is_member,
        CASE WHEN g.group_owner = $1 THEN TRUE ELSE FALSE END AS is_owner,
        COALESCE(gm.joined_at, g.created_at) AS sort_date
    FROM groups g
    LEFT JOIN group_members gm
        ON gm.group_id = g.id
        AND gm.user_id = $1
        AND gm.deleted_at IS NULL
    WHERE g.deleted_at IS NULL
      AND (gm.user_id = $1 OR g.group_owner = $1)
) t
ORDER BY sort_date DESC, group_id DESC
LIMIT $2 OFFSET $3
`

type GetUserGroupsParams struct {
	GroupOwner int64
	Limit      int32
	Offset     int32
}

type GetUserGroupsRow struct {
	GroupID          int64
	GroupOwner       int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
	MembersCount     int32
	IsMember         bool
	IsOwner          bool
}

func (q *Queries) GetUserGroups(ctx context.Context, arg GetUserGroupsParams) ([]GetUserGroupsRow, error) {
	rows, err := q.db.Query(ctx, getUserGroups, arg.GroupOwner, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetUserGroupsRow{}
	for rows.Next() {
		var i GetUserGroupsRow
		if err := rows.Scan(
			&i.GroupID,
			&i.GroupOwner,
			&i.GroupTitle,
			&i.GroupDescription,
			&i.GroupImageID,
			&i.MembersCount,
			&i.IsMember,
			&i.IsOwner,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const isGroupMembershipPending = `-- name: GetPendingGroupMembershipState :one
SELECT
    EXISTS (
        SELECT 1
        FROM group_join_requests gjr
        WHERE gjr.group_id = $1
          AND gjr.user_id = $2
          AND gjr.status = 'pending'
          AND gjr.deleted_at IS NULL
    ) AS has_pending_join_request,

    EXISTS (
        SELECT 1
        FROM group_invites gi
        WHERE gi.group_id = $1
          AND gi.receiver_id = $2
          AND gi.status = 'pending'
          AND gi.deleted_at IS NULL
    ) AS has_pending_invite;
`

type IsGroupMembershipPendingParams struct {
	GroupID int64
	UserID  int64
}

type IsGroupMembershipPendingRow struct {
	HasPendingJoinRequest bool
	HasPendingInvite      bool
}

func (q *Queries) IsGroupMembershipPending(
	ctx context.Context,
	arg IsGroupMembershipPendingParams,
) (IsGroupMembershipPendingRow, error) {

	row := q.db.QueryRow(ctx, isGroupMembershipPending, arg.GroupID, arg.UserID)

	var result IsGroupMembershipPendingRow
	err := row.Scan(
		&result.HasPendingJoinRequest,
		&result.HasPendingInvite,
	)

	return result, err
}

const isUserGroupMember = `-- name: IsUserGroupMember :one
SELECT EXISTS (
    SELECT 1
    FROM group_members
    WHERE group_id = $1
      AND user_id = $2
      AND deleted_at IS NULL
) AS is_member
`

type IsUserGroupMemberParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) IsUserGroupMember(ctx context.Context, arg IsUserGroupMemberParams) (bool, error) {
	row := q.db.QueryRow(ctx, isUserGroupMember, arg.GroupID, arg.UserID)
	var is_member bool
	err := row.Scan(&is_member)
	return is_member, err
}

const isUserGroupOwner = `-- name: IsUserGroupOwner :one
SELECT (group_owner = $2) AS is_owner
FROM groups
WHERE id = $1
  AND deleted_at IS NULL
`

type IsUserGroupOwnerParams struct {
	ID         int64
	GroupOwner int64
}

func (q *Queries) IsUserGroupOwner(ctx context.Context, arg IsUserGroupOwnerParams) (bool, error) {
	row := q.db.QueryRow(ctx, isUserGroupOwner, arg.ID, arg.GroupOwner)
	var is_owner bool
	err := row.Scan(&is_owner)
	return is_owner, err
}

const leaveGroup = `-- name: LeaveGroup :exec
UPDATE group_members
SET deleted_at = CURRENT_TIMESTAMP
WHERE group_id = $1
  AND user_id = $2
  AND role <> 'owner'
`

type LeaveGroupParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) LeaveGroup(ctx context.Context, arg LeaveGroupParams) error {
	_, err := q.db.Exec(ctx, leaveGroup, arg.GroupID, arg.UserID)
	return err
}

const rejectGroupJoinRequest = `-- name: RejectGroupJoinRequest :exec
UPDATE group_join_requests
SET status = 'rejected'
WHERE group_id = $1
  AND user_id = $2
`

type RejectGroupJoinRequestParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) RejectGroupJoinRequest(ctx context.Context, arg RejectGroupJoinRequestParams) error {
	_, err := q.db.Exec(ctx, rejectGroupJoinRequest, arg.GroupID, arg.UserID)
	return err
}

const searchGroups = `-- name: SearchGroups :many
SELECT DISTINCT ON (g.id)
    g.id,
    g.group_owner,
    g.group_title,
    g.group_description,
    g.group_image_id,
    g.members_count,
    (gm.user_id IS NOT NULL) AS is_member,
    (g.group_owner = $2) AS is_owner,
    CASE
        WHEN LENGTH($1) >= 3 THEN
            similarity(g.group_title, $1) * 2.0 +
            similarity(g.group_description, $1)
        ELSE
            CASE
                WHEN g.group_title ILIKE '%' || $1 || '%' THEN 3
                WHEN g.group_description ILIKE '%' || $1 || '%' THEN 1
                ELSE 0
            END
    END AS weighted_score
FROM groups g
LEFT JOIN group_members gm
    ON gm.group_id = g.id
   AND gm.user_id = $2
   AND gm.deleted_at IS NULL
WHERE g.deleted_at IS NULL
  AND (
        -- Always allow substring match for any length
        g.group_title ILIKE '%' || $1 || '%'
     OR g.group_description ILIKE '%' || $1 || '%'
     -- For longer queries, also use fuzzy % match
     OR (LENGTH($1) >= 3 AND (g.group_title % $1 OR g.group_description % $1))
      )
ORDER BY
    g.id,                    -- required for DISTINCT ON
    weighted_score DESC,      -- sort by relevance
    (gm.user_id IS NOT NULL) DESC, -- prioritize groups the user belongs to
    g.members_count DESC,     -- then by members count
    g.id DESC                -- stable tie-breaker
LIMIT $3 OFFSET $4;
`

type SearchGroupsParams struct {
	Query  string
	UserID int64
	Limit  int32
	Offset int32
}

type SearchGroupsRow struct {
	ID               int64
	GroupOwner       int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
	MembersCount     int32
	IsMember         bool
	IsOwner          bool
	WeightedScore    float64
}

func (q *Queries) SearchGroups(
	ctx context.Context,
	arg SearchGroupsParams,
) ([]SearchGroupsRow, error) {

	rows, err := q.db.Query(ctx,
		searchGroups,
		arg.Query,
		arg.UserID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []SearchGroupsRow{}
	for rows.Next() {
		var i SearchGroupsRow
		if err := rows.Scan(
			&i.ID,
			&i.GroupOwner,
			&i.GroupTitle,
			&i.GroupDescription,
			&i.GroupImageID,
			&i.MembersCount,
			&i.IsMember,
			&i.IsOwner,
			&i.WeightedScore,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

const sendGroupInvites = `-- name: SendGroupInvite :exec
INSERT INTO group_invites (group_id, sender_id, receiver_id, status)
SELECT
    $1 AS group_id,
    $2 AS sender_id,
    receiver_id,
    'pending'
FROM unnest($3::bigint[]) AS receiver_id
ON CONFLICT (group_id, receiver_id)
DO UPDATE SET status = 'pending';
`

type SendGroupInvitesParams struct {
	GroupID     int64
	SenderID    int64
	ReceiverIDs []int64
}

func (q *Queries) SendGroupInvites(ctx context.Context, arg SendGroupInvitesParams) error {
	_, err := q.db.Exec(
		ctx,
		sendGroupInvites,
		arg.GroupID,
		arg.SenderID,
		arg.ReceiverIDs,
	)
	return err
}

const getGroupInviterId = `--name: GetGroupInviterId :one
SELECT sender_id
FROM group_invites
WHERE group_id = $1
  AND receiver_id = $2
  AND status = 'pending'
  AND deleted_at IS NULL;
  `

type GetGroupInviterIdParams struct {
	GroupID    int64
	ReceiverID int64
}

func (q *Queries) GetGroupInviterId(ctx context.Context, arg GetGroupInviterIdParams) (int64, error) {
	row := q.db.QueryRow(ctx, getGroupInviterId,
		arg.GroupID,
		arg.ReceiverID,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const sendGroupJoinRequest = `-- name: SendGroupJoinRequest :exec
INSERT INTO group_join_requests (group_id, user_id, status)
VALUES ($1, $2, 'pending')
ON CONFLICT (group_id, user_id)
DO UPDATE SET status = 'pending'
`

type SendGroupJoinRequestParams struct {
	GroupID int64
	UserID  int64
}

func (q *Queries) SendGroupJoinRequest(ctx context.Context, arg SendGroupJoinRequestParams) error {
	_, err := q.db.Exec(ctx, sendGroupJoinRequest, arg.GroupID, arg.UserID)
	return err
}

const softDeleteGroup = `-- name: SoftDeleteGroup :exec
UPDATE groups
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1
`

func (q *Queries) SoftDeleteGroup(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, softDeleteGroup, id)
	return err
}

const transferOwnership = `-- name: TransferOwnership :exec


WITH demote AS (
    UPDATE group_members AS gm_old
    SET role = 'member'
    WHERE gm_old.group_id = $1
      AND gm_old.user_id = $2
      AND gm_old.role = 'owner'
),
promote AS (
    UPDATE group_members AS gm_new
    SET role = 'owner'
    WHERE gm_new.group_id = $1
      AND gm_new.user_id = $3
      AND gm_new.role = 'member'
)
SELECT 1
`

type TransferOwnershipParams struct {
	GroupID  int64
	UserID   int64
	UserID_2 int64
}

// owners cannot leave the group (transfer ownership logic? TODO)
func (q *Queries) TransferOwnership(ctx context.Context, arg TransferOwnershipParams) error {
	_, err := q.db.Exec(ctx, transferOwnership, arg.GroupID, arg.UserID, arg.UserID_2)
	return err
}

const updateGroup = `-- name: UpdateGroup :execrows
UPDATE groups
SET
    group_title      = $2,
    group_description    = $3,
    group_image_id     = $4
WHERE id = $1 AND deleted_at IS NULL
`

type UpdateGroupParams struct {
	ID               int64
	GroupTitle       string
	GroupDescription string
	GroupImageID     int64
}

func (q *Queries) UpdateGroup(ctx context.Context, arg UpdateGroupParams) (int64, error) {
	result, err := q.db.Exec(ctx, updateGroup,
		arg.ID,
		arg.GroupTitle,
		arg.GroupDescription,
		arg.GroupImageID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

const userGroupCountsPerRole = `-- name: UserGroupCountsPerRole :one
SELECT
    COUNT(*) FILTER (WHERE g.group_owner = $1) AS owner_count,
    COUNT(*) FILTER (WHERE gm.role = 'member' AND g.group_owner <> $1) AS member_only_count,
    COUNT(*) AS total_memberships
FROM group_members gm
JOIN groups g ON gm.group_id = g.id
WHERE gm.user_id = $1
  AND gm.deleted_at IS NULL
  AND g.deleted_at IS NULL
`

type UserGroupCountsPerRoleRow struct {
	OwnerCount       int64
	MemberOnlyCount  int64
	TotalMemberships int64
}

func (q *Queries) UserGroupCountsPerRole(ctx context.Context, groupOwner int64) (UserGroupCountsPerRoleRow, error) {
	row := q.db.QueryRow(ctx, userGroupCountsPerRole, groupOwner)
	var i UserGroupCountsPerRoleRow
	err := row.Scan(&i.OwnerCount, &i.MemberOnlyCount, &i.TotalMemberships)
	return i, err
}

const getFollowersNotInvitedToGroup = `-- name: GetFollowersNotInvitedToGroup :many
SELECT
    u.id,
    u.username,
    u.avatar_id
FROM follows f
JOIN users u
    ON u.id = f.follower_id
LEFT JOIN group_invites gi
    ON gi.group_id = $2
   AND gi.receiver_id = u.id
   AND gi.deleted_at IS NULL
WHERE f.following_id = $1          -- people who follow the given user
  AND f.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND u.current_status = 'active'
  AND gi.receiver_id IS NULL      -- exclude anyone already invited
  ORDER BY f.created_at DESC
  LIMIT $3 OFFSET $4;
  `

type GetFollowersNotInvitedToGroupParams struct {
	UserId  int64
	GroupId int64
	Limit   int
	Offset  int
}

type GetFollowersNotInvitedToGroupRow struct {
	Id       int64
	Username string
	AvatarId int64
}

func (q *Queries) GetFollowersNotInvitedToGroup(ctx context.Context, arg GetFollowersNotInvitedToGroupParams) ([]GetFollowersNotInvitedToGroupRow, error) {
	rows, err := q.db.Query(ctx, getFollowersNotInvitedToGroup, arg.UserId, arg.GroupId, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetFollowersNotInvitedToGroupRow{}
	for rows.Next() {
		var i GetFollowersNotInvitedToGroupRow
		if err := rows.Scan(
			&i.Id,
			&i.Username,
			&i.AvatarId,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPendingGroupJoinRequests = `-- name: GetPendingGroupJoinRequests :many
SELECT
    u.id,
    u.username,
    u.avatar_id
FROM group_join_requests gjr
JOIN users u
    ON u.id = gjr.user_id
WHERE gjr.group_id = $1
  AND gjr.status = 'pending'
  AND gjr.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND u.current_status = 'active'
ORDER BY gjr.created_at ASC
LIMIT $2 OFFSET $3;
`

type GetPendingGroupJoinRequestsParams struct {
	GroupId int64
	Limit   int
	Offset  int
}

type GetPendingGroupJoinRequestsRow struct {
	Id       int64
	Username string
	AvatarId int64
}

func (q *Queries) GetPendingGroupJoinRequests(ctx context.Context, arg GetPendingGroupJoinRequestsParams) ([]GetPendingGroupJoinRequestsRow, error) {
	rows, err := q.db.Query(ctx, getPendingGroupJoinRequests, arg.GroupId, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []GetPendingGroupJoinRequestsRow{}
	for rows.Next() {
		var i GetPendingGroupJoinRequestsRow
		if err := rows.Scan(
			&i.Id,
			&i.Username,
			&i.AvatarId,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPendingGroupJoinRequestsCount = `-- name: GetPendingGroupJoinRequestsCount :one
SELECT COUNT(*)
FROM group_join_requests gjr
JOIN users u
    ON u.id = gjr.user_id
WHERE gjr.group_id = $1
  AND gjr.status = 'pending'
  AND gjr.deleted_at IS NULL
  AND u.deleted_at IS NULL
  AND u.current_status = 'active';
`

type GetPendingGroupJoinRequestsCountParams struct {
	GroupId int64
}

func (q *Queries) GetPendingGroupJoinRequestsCount(
	ctx context.Context,
	arg GetPendingGroupJoinRequestsCountParams,
) (int64, error) {
	var count int64
	err := q.db.QueryRow(
		ctx,
		getPendingGroupJoinRequestsCount,
		arg.GroupId,
	).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
