package dbservice

import (
	"context"
)

type Querier interface {
	AcceptFollowRequest(ctx context.Context, arg AcceptFollowRequestParams) error
	AcceptGroupInvite(ctx context.Context, arg AcceptGroupInviteParams) error
	AcceptGroupJoinRequest(ctx context.Context, arg AcceptGroupJoinRequestParams) error
	AddGroupOwnerAsMember(ctx context.Context, arg AddGroupOwnerAsMemberParams) error
	AddUserToGroup(ctx context.Context, arg AddUserToGroupParams) error
	AreFollowingEachOther(ctx context.Context, arg AreFollowingEachOtherParams) (AreFollowingEachOtherRow, error)
	BanUser(ctx context.Context, arg BanUserParams) error
	CancelGroupInvite(ctx context.Context, arg CancelGroupInviteParams) error
	CancelGroupJoinRequest(ctx context.Context, arg CancelGroupJoinRequestParams) error
	CreateGroup(ctx context.Context, arg CreateGroupParams) (int64, error)
	DeclineGroupInvite(ctx context.Context, arg DeclineGroupInviteParams) error
	FollowUser(ctx context.Context, arg FollowUserParams) (string, error)
	GetAllGroups(ctx context.Context, arg GetAllGroupsParams) ([]GetAllGroupsRow, error)
	GetAllGroupMemberIds(ctx context.Context, arg GetAllGroupMemberIdsParams) ([]GetAllGroupMemberIdsRow, error)
	GetBatchUsersBasic(ctx context.Context, dollar_1 []int64) ([]GetBatchUsersBasicRow, error)
	// S1: second-degree follows
	// S2: shared groups
	// Combine & score
	GetFollowSuggestions(ctx context.Context, followerID int64) ([]GetFollowSuggestionsRow, error)
	GetFollowerCount(ctx context.Context, followingID int64) (int64, error)
	GetFollowers(ctx context.Context, arg GetFollowersParams) ([]GetFollowersRow, error)
	GetFollowersNotInvitedToGroup(ctx context.Context, arg GetFollowersNotInvitedToGroupParams) ([]GetFollowersNotInvitedToGroupRow, error)
	GetFollowing(ctx context.Context, arg GetFollowingParams) ([]GetFollowingRow, error)
	GetFollowingCount(ctx context.Context, followerID int64) (int64, error)
	GetFollowingIds(ctx context.Context, followerID int64) ([]int64, error)
	GetGroupInfo(ctx context.Context, id int64) (GetGroupInfoRow, error)
	GetGroupInviterId(ctx context.Context, arg GetGroupInviterIdParams) (int64, error)
	GetGroupBasicInfo(ctx context.Context, id int64) (GetGroupBasicInfoRow, error)
	GetGroupMembers(ctx context.Context, arg GetGroupMembersParams) ([]GetGroupMembersRow, error)
	GetMutualFollowers(ctx context.Context, arg GetMutualFollowersParams) ([]GetMutualFollowersRow, error)
	GetPendingGroupJoinRequests(ctx context.Context, arg GetPendingGroupJoinRequestsParams) ([]GetPendingGroupJoinRequestsRow, error)
	GetPendingGroupJoinRequestsCount(ctx context.Context, arg GetPendingGroupJoinRequestsCountParams) (int64, error)
	GetUserBasic(ctx context.Context, id int64) (GetUserBasicRow, error)
	GetUserForLogin(ctx context.Context, arg GetUserForLoginParams) (GetUserForLoginRow, error)
	GetUserGroupRole(ctx context.Context, arg GetUserGroupRoleParams) (NullGroupRole, error)
	GetUserGroups(ctx context.Context, arg GetUserGroupsParams) ([]GetUserGroupsRow, error)
	GetUserPassword(ctx context.Context, userID int64) (string, error)
	GetUserProfile(ctx context.Context, id int64) (GetUserProfileRow, error)
	InsertNewUser(ctx context.Context, arg InsertNewUserParams) (int64, error)
	InsertNewUserAuth(ctx context.Context, arg InsertNewUserAuthParams) error
	IsFollowRequestPending(ctx context.Context, arg IsFollowRequestPendingParams) (bool, error)
	IsFollowing(ctx context.Context, arg IsFollowingParams) (bool, error)
	IsGroupMembershipPending(ctx context.Context, arg IsGroupMembershipPendingParams) (IsGroupMembershipPendingRow, error)
	IsUserGroupMember(ctx context.Context, arg IsUserGroupMemberParams) (bool, error)
	IsUserGroupOwner(ctx context.Context, arg IsUserGroupOwnerParams) (bool, error)
	LeaveGroup(ctx context.Context, arg LeaveGroupParams) error
	RejectFollowRequest(ctx context.Context, arg RejectFollowRequestParams) error
	RejectGroupJoinRequest(ctx context.Context, arg RejectGroupJoinRequestParams) error
	RemoveImages(ctx context.Context, arg []int64) error
	SearchGroups(ctx context.Context, arg SearchGroupsParams) ([]SearchGroupsRow, error)
	SearchUsers(ctx context.Context, arg SearchUsersParams) ([]SearchUsersRow, error)
	SendGroupInvites(ctx context.Context, arg SendGroupInvitesParams) error
	SendGroupJoinRequest(ctx context.Context, arg SendGroupJoinRequestParams) error
	SoftDeleteGroup(ctx context.Context, id int64) error
	SoftDeleteUser(ctx context.Context, id int64) error
	// owners cannot leave the group (transfer ownership logic? TODO)
	TransferOwnership(ctx context.Context, arg TransferOwnershipParams) error
	UnbanUser(ctx context.Context, id int64) error
	//1: follower_id
	//2: following_id
	// returns followed or requested depending on target's privacy settings
	UnfollowUser(ctx context.Context, arg UnfollowUserParams) (string, error)
	UpdateGroup(ctx context.Context, arg UpdateGroupParams) (int64, error)
	UpdateProfilePrivacy(ctx context.Context, arg UpdateProfilePrivacyParams) error
	UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) error
	UpdateUserPassword(ctx context.Context, arg UpdateUserPasswordParams) error
	UpdateUserProfile(ctx context.Context, arg UpdateUserProfileParams) (User, error)
	UserGroupCountsPerRole(ctx context.Context, groupOwner int64) (UserGroupCountsPerRoleRow, error)
}

var _ Querier = (*Queries)(nil)
