package models

import (
	ct "social-network/shared/go/ct"
)

//-------------------------------------------
// Auth
//-------------------------------------------

type RegisterUserRequest struct {
	Username    ct.Username       `json:"username" validate:"nullable"`
	FirstName   ct.Name           `json:"first_name"`
	LastName    ct.Name           `json:"last_name"`
	DateOfBirth ct.DateOfBirth    `json:"date_of_birth"`
	AvatarId    ct.Id             `json:"avatar_id" validate:"nullable"`
	About       ct.About          `json:"about" validate:"nullable"`
	Public      bool              `json:"public"`
	Email       ct.Email          `json:"email"`
	Password    ct.HashedPassword `json:"password"`
}

type RegisterUserResponse struct {
	UserId   int64       `json:"user_id"`
	Username ct.Username `json:"username"`
}

type LoginRequest struct {
	Identifier ct.Identifier     `json:"identifier"`
	Password   ct.HashedPassword `json:"password"`
}

type UpdatePasswordRequest struct {
	UserId      ct.Id             `json:"user_id"`
	OldPassword ct.HashedPassword `json:"old_password"`
	NewPassword ct.HashedPassword `json:"new_password"`
}

type UpdateEmailRequest struct {
	UserId ct.Id    `json:"user_id"`
	Email  ct.Email `json:"email"`
}

//-------------------------------------------
// Profile
//-------------------------------------------

type UserId int64

type User struct {
	UserId    ct.Id       `json:"id"`
	Username  ct.Username `json:"username"`
	AvatarId  ct.Id       `json:"avatar_id" validate:"nullable"`
	AvatarURL string      `json:"avatar_url"`
}

type Users struct {
	Users []User `json:"users"`
}

type UserSearchReq struct {
	SearchTerm ct.SearchTerm `json:"search_term"`
	Limit      ct.Limit      `json:"limit"`
}

type UserProfileRequest struct {
	UserId      ct.Id `json:"user_id"`
	RequesterId ct.Id `json:"requester_id"`
}

type UserProfileResponse struct {
	UserId                        ct.Id          `json:"user_id"`
	Username                      ct.Username    `json:"username"`
	FirstName                     ct.Name        `json:"first_name"`
	LastName                      ct.Name        `json:"last_name"`
	DateOfBirth                   ct.DateOfBirth `json:"date_of_birth"`
	AvatarId                      ct.Id          `json:"avatar_id" validate:"nullable"`
	AvatarURL                     string         `json:"avatar_url"`
	About                         ct.About       `json:"about"`
	Public                        bool           `json:"public"`
	CreatedAt                     ct.GenDateTime `json:"created_at"`
	Email                         ct.Email       `json:"email"`
	FollowersCount                int64          `json:"followers_count"`
	FollowingCount                int64          `json:"following_count"`
	GroupsCount                   int64          `json:"groups_count"`
	OwnedGroupsCount              int64          `json:"owned_groups_count"`
	ViewerIsFollowing             bool           `json:"viewer_is_following"`
	OwnProfile                    bool           `json:"own_profile"`
	IsPending                     bool           `json:"is_pending"`
	FollowRequestFromProfileOwner bool           `json:"follow_request_from_profile_owner"`
}

type UpdateProfileRequest struct {
	UserId      ct.Id
	Username    ct.Username    `json:"username"`
	FirstName   ct.Name        `json:"first_name"`
	LastName    ct.Name        `json:"last_name"`
	DateOfBirth ct.DateOfBirth `json:"date_of_birth"`
	AvatarId    ct.Id          `json:"avatar_id" validate:"nullable"`
	About       ct.About       `json:"about" validate:"nullable"`
	DeleteImage bool           `json:"delete_image"`
}

type UpdateProfilePrivacyRequest struct {
	UserId ct.Id `json:"user_id"`
	Public bool  `json:"public"`
}

// -------------------------------------------
// Groups
// -------------------------------------------

type GroupId ct.Id

type GroupRole string

type GroupMembersReq struct {
	UserId  ct.Id     `json:"user_id"`
	GroupId ct.Id     `json:"group_id"`
	Limit   ct.Limit  `json:"limit"`
	Offset  ct.Offset `json:"offset"`
}

type Pagination struct {
	UserId ct.Id     `json:"user_id"`
	Limit  ct.Limit  `json:"limit"`
	Offset ct.Offset `json:"offset"`
}

type GroupUser struct {
	UserId    ct.Id       `json:"user_id"`
	Username  ct.Username `json:"username"`
	AvatarId  ct.Id       `json:"avatar_id" validate:"nullable"`
	AvatarUrl string      `json:"avatar_url"`
	GroupRole string      `json:"group_role"`
}

type GroupUsers struct {
	GroupUsers []GroupUser `json:"group_users"`
}

type GroupSearchReq struct {
	SearchTerm ct.SearchTerm `json:"search_term"`
	UserId     ct.Id         `json:"user_id"`
	Limit      ct.Limit      `json:"limit"`
	Offset     ct.Offset     `json:"offset"`
}

type Group struct {
	GroupId          ct.Id    `json:"group_id"`
	GroupOwnerId     ct.Id    `json:"group_owner_id"`
	GroupTitle       ct.Title `json:"group_title"`
	GroupDescription ct.About `json:"group_description"`
	GroupImage       ct.Id    `json:"group_image_id" validate:"nullable"`
	GroupImageURL    string   `json:"group_image_url"`
	MembersCount     int32    `json:"members_count"`
	IsMember         bool     `json:"is_member"`
	IsOwner          bool     `json:"is_owner"`
	PendingRequest   bool     `json:"pending_request"`
	PendingInvite    bool     `json:"pending_invite"`
}

type Groups struct {
	Groups []Group `json:"groups"`
}

type InviteToGroupReq struct {
	InviterId  ct.Id
	InvitedIds ct.Ids `json:"invited_id"`
	GroupId    ct.Id  `json:"group_id"`
}

type HandleGroupInviteRequest struct {
	GroupId   ct.Id `json:"group_id"`
	InvitedId ct.Id `json:"invited_id"`
	Accepted  bool  `json:"accepted"`
}

type GroupJoinRequest struct {
	GroupId     ct.Id `json:"group_id"`
	RequesterId ct.Id
}

type HandleJoinRequest struct {
	GroupId     ct.Id `json:"group_id"`
	RequesterId ct.Id `json:"requester_id"`
	OwnerId     ct.Id `json:"owner_id"`
	Accepted    bool  `json:"accepted"`
}

type GeneralGroupReq struct {
	GroupId ct.Id `json:"group_id"`
	UserId  ct.Id `json:"user_id"`
}

type RemoveFromGroupRequest struct {
	GroupId  ct.Id `json:"group_id"`
	MemberId ct.Id `json:"member_id"`
	OwnerId  ct.Id `json:"owner_id"`
}

type CreateGroupRequest struct {
	OwnerId          ct.Id    `json:"owner_id"`
	GroupTitle       ct.Title `json:"group_title"`
	GroupDescription ct.About `json:"group_description"`
	GroupImage       ct.Id    `json:"group_image_id" validate:"nullable"`
}

type UpdateGroupRequest struct {
	RequesterId      ct.Id
	GroupId          ct.Id    `json:"group_id"`
	GroupTitle       ct.Title `json:"group_title"`
	GroupDescription ct.About `json:"group_description"`
	GroupImage       ct.Id    `json:"group_image_id" validate:"nullable"`
	DeleteImage      bool     `json:"delete_image"`
}

// -------------------------------------------
// Followers
// -------------------------------------------

type FollowUserReq struct {
	FollowerId   ct.Id
	TargetUserId ct.Id `json:"target_user_id"`
}

type FollowUserResp struct {
	IsPending         bool `json:"is_pending"`
	ViewerIsFollowing bool `json:"viewer_is_following"`
}

type HandleFollowRequestReq struct {
	UserId      ct.Id `json:"user_id"`
	RequesterId ct.Id `json:"requester_id"`
	Accept      bool  `json:"accept"`
}

type FollowRelationship struct {
	FollowerFollowsTarget bool `json:"follower_follows_target"`
	TargetFollowsFollower bool `json:"target_follows_follower"`
}
