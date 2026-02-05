/*
Expose methods via gRpc
*/

package handler

import (
	"context"
	"fmt"
	"runtime"
	cm "social-network/shared/gen-go/common"
	pb "social-network/shared/gen-go/users"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// AUTH
func (s *UsersHandler) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	tele.Info(ctx, "RegisterUser gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	resp, err := s.Application.RegisterUser(ctx, models.RegisterUserRequest{
		Username:    ct.Username(req.GetUsername()),
		FirstName:   ct.Name(req.GetFirstName()),
		LastName:    ct.Name(req.GetLastName()),
		Email:       ct.Email(req.GetEmail()),
		Password:    ct.HashedPassword(req.GetPassword()),
		DateOfBirth: ct.DateOfBirth(req.GetDateOfBirth().AsTime()),
		AvatarId:    ct.Id(req.GetAvatar()),
		About:       ct.About(req.GetAbout()),
		Public:      req.GetPublic(),
	})
	if err != nil {
		tele.Warn(ctx, "Error in RegisterUser. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &pb.RegisterUserResponse{
		UserId:   resp.UserId,
		Username: resp.Username.String(),
	}, nil
}

func (s *UsersHandler) LoginUser(ctx context.Context, req *pb.LoginRequest) (*cm.User, error) {
	tele.Info(ctx, "LoginUser gRPC method called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "LoginUser: request is nil")
	}

	Identifier := req.GetIdentifier()
	if err := invalidString("ident", Identifier); err != nil {
		return nil, err
	}

	Password := req.GetPassword()
	if err := invalidString("pass", Password); err != nil {
		return nil, err
	}

	user, err := s.Application.LoginUser(ctx, models.LoginRequest{
		Identifier: ct.Identifier(Identifier),
		Password:   ct.HashedPassword(Password),
	})
	if err != nil {
		tele.Warn(ctx, "Error in LoginUser. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &cm.User{
		UserId:    user.UserId.Int64(),
		Username:  user.Username.String(),
		Avatar:    user.AvatarId.Int64(),
		AvatarUrl: user.AvatarURL,
	}, nil
}

func (s *UsersHandler) UpdateUserPassword(ctx context.Context, req *pb.UpdatePasswordRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "UpdateUserPassword gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "UpdateUserPassword: request is nil")
	}

	// userId := req.GetUserId()
	// if err := invalidId("userId", userId); err != nil {
	// 	return nil, err
	// }

	// newPassword := req.GetNewPassword()
	// if err := invalidString("newPassword", newPassword); err != nil {
	// 	return nil, err
	// }

	err := s.Application.UpdateUserPassword(ctx, models.UpdatePasswordRequest{
		UserId:      ct.Id(req.UserId),
		OldPassword: ct.HashedPassword(req.OldPassword),
		NewPassword: ct.HashedPassword(req.NewPassword),
	})
	if err != nil {
		tele.Warn(ctx, "Error in UpdateUserPassword. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) UpdateUserEmail(ctx context.Context, req *pb.UpdateEmailRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "UpdateUserEmail gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "UpdateUserEmail: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("user id", userId); err != nil {
		return nil, err
	}

	newEmail := req.GetEmail()
	if err := invalidString("newEmail", newEmail); err != nil {
		return nil, err
	}

	err := s.Application.UpdateUserEmail(ctx, models.UpdateEmailRequest{
		UserId: ct.Id(userId),
		Email:  ct.Email(newEmail),
	})
	if err != nil {
		tele.Warn(ctx, "Error in UpdateUserEmail. @1", "error", err.Error(), "req", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

// FOLLOW
func (s *UsersHandler) GetFollowersPaginated(ctx context.Context, req *pb.Pagination) (*cm.ListUsers, error) {
	tele.Info(ctx, "GetFollowersPaginated gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetFollowersPaginated: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	offset := req.GetOffset()
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	pag := models.Pagination{
		UserId: ct.Id(userId),
		Limit:  ct.Limit(limit),
		Offset: ct.Offset(offset),
	}

	resp, err := s.Application.GetFollowersPaginated(ctx, pag)
	if err != nil {
		tele.Error(ctx, "Error in GetFollowersPaginated. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return usersToPB(resp), nil
}

func (s *UsersHandler) GetFollowingPaginated(ctx context.Context, req *pb.Pagination) (*cm.ListUsers, error) {
	tele.Info(ctx, "GetFollowingPaginated gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetFollowingPaginated: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	offset := req.GetOffset()
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	pag := models.Pagination{
		UserId: ct.Id(userId),
		Limit:  ct.Limit(limit),
		Offset: ct.Offset(offset),
	}

	resp, err := s.Application.GetFollowingPaginated(ctx, pag)
	if err != nil {
		tele.Warn(ctx, "Error in GetFollowingPaginated. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return usersToPB(resp), nil
}

func (s *UsersHandler) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.FollowUserResponse, error) {
	tele.Info(ctx, "FollowUser gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "FollowUser: request is nil")
	}

	followerId := req.GetFollowerId()
	if err := invalidId("followerId", followerId); err != nil {
		return nil, err
	}

	targetUserId := req.GetTargetUserId()
	if err := invalidId("targetUserId", targetUserId); err != nil {
		return nil, err
	}

	resp, err := s.Application.FollowUser(ctx, models.FollowUserReq{
		FollowerId:   ct.Id(followerId),
		TargetUserId: ct.Id(targetUserId),
	})
	if err != nil {
		tele.Error(ctx, "Error in FollowUser. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &pb.FollowUserResponse{
		IsPending:         resp.IsPending,
		ViewerIsFollowing: resp.ViewerIsFollowing,
	}, nil
}

func (s *UsersHandler) UnFollowUser(ctx context.Context, req *pb.FollowUserRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "UnFollowUser gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "UnFollowUser: request is nil")
	}

	followerId := req.GetFollowerId()
	if err := invalidId("followerId", followerId); err != nil {
		return nil, err
	}

	targetUserId := req.GetTargetUserId()
	if err := invalidId("targetUserId", targetUserId); err != nil {
		return nil, err
	}

	err := s.Application.UnFollowUser(ctx, models.FollowUserReq{
		FollowerId:   ct.Id(followerId),
		TargetUserId: ct.Id(targetUserId),
	})
	if err != nil {
		tele.Warn(ctx, "Error in UnFollowUser. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) HandleFollowRequest(ctx context.Context, req *pb.HandleFollowRequestRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "HandleFollowRequest gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "HandleFollowRequest: request is nil")
	}

	userID := req.GetUserId()
	if err := invalidId("userId", userID); err != nil {
		return nil, err
	}

	RequesterId := req.GetRequesterId()
	if err := invalidId("requesterId", RequesterId); err != nil {
		return nil, err
	}

	acc := req.GetAccept()

	err := s.Application.HandleFollowRequest(ctx, models.HandleFollowRequestReq{
		UserId:      ct.Id(userID),
		RequesterId: ct.Id(RequesterId),
		Accept:      acc,
	})
	if err != nil {
		tele.Error(ctx, "Error in HandleFollowRequest. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) GetFollowingIds(ctx context.Context, req *wrapperspb.Int64Value) (*cm.UserIds, error) {
	tele.Info(ctx, "GetFollowingIds gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetFollowingIds: request is nil")
	}
	userId := req.GetValue()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	resp, err := s.Application.GetFollowingIds(ctx, ct.Id(userId))
	if err != nil {
		tele.Error(ctx, "Error in GetFollowingIds. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &cm.UserIds{Values: resp}, nil
}

func (s *UsersHandler) GetFollowSuggestions(ctx context.Context, req *wrapperspb.Int64Value) (*cm.ListUsers, error) {
	tele.Info(ctx, "GetFollowSuggestions gRPC method called with @1", "request", req.String())
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetFollowSuggestions: request is nil")
	}

	userId := req.GetValue()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	resp, err := s.Application.GetFollowSuggestions(ctx, ct.Id(userId))
	if err != nil {
		tele.Error(ctx, "Error in GetFollowSuggestions. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return usersToPB(resp), nil
}

func (s *UsersHandler) IsFollowing(ctx context.Context, req *pb.IsFollowingRequest) (*wrapperspb.BoolValue, error) {
	tele.Info(ctx, "IsFollowing called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "IsFollowing: request is nil")
	}

	followerId := req.GetFollowerId()
	if err := invalidId("followerId", followerId); err != nil {
		return nil, err
	}

	targetUserId := req.GetTargetUserId()
	if err := invalidId("targetUserId", targetUserId); err != nil {
		return nil, err
	}

	resp, err := s.Application.IsFollowing(ctx, models.FollowUserReq{
		FollowerId:   ct.Id(followerId),
		TargetUserId: ct.Id(targetUserId),
	})
	if err != nil {
		tele.Error(ctx, "Error in IsFollowing. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return wrapperspb.Bool(resp), nil
}

func (s *UsersHandler) AreFollowingEachOther(ctx context.Context, req *pb.FollowUserRequest) (*pb.AreFollowingEachOtherResponse, error) {
	tele.Info(ctx, "AreFollowingEachOther called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "AreFollowingEachOther: request is nil")
	}

	userAId := req.GetFollowerId()
	if err := invalidId("userAId", userAId); err != nil {
		return nil, err
	}

	userBId := req.GetTargetUserId()
	if err := invalidId("userBId", userBId); err != nil {
		return nil, err
	}

	resp, err := s.Application.AreFollowingEachOther(ctx, models.FollowUserReq{
		FollowerId:   ct.Id(userAId),
		TargetUserId: ct.Id(userBId),
	})
	if err != nil {
		tele.Error(ctx, "Error in AreFollowingEachOther. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &pb.AreFollowingEachOtherResponse{
			FollowerFollowsTarget: resp.FollowerFollowsTarget,
			TargetFollowsFollower: resp.TargetFollowsFollower,
		},
		nil
}

// GROUPS
func (s *UsersHandler) GetAllGroupsPaginated(ctx context.Context, req *pb.Pagination) (*pb.GroupArr, error) {
	tele.Info(ctx, "GetAllGroupsPaginated called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetAllGroupsPaginated: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	offset := req.GetOffset()
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	pag := models.Pagination{
		UserId: ct.Id(userId),
		Limit:  ct.Limit(limit),
		Offset: ct.Offset(offset),
	}

	resp, err := s.Application.GetAllGroupsPaginated(ctx, pag)
	if err != nil {
		tele.Error(ctx, "Error in GetAllGroupsPaginated. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return groupsToPb(resp), nil
}

func (s *UsersHandler) GetUserGroupsPaginated(ctx context.Context, req *pb.Pagination) (*pb.GroupArr, error) {
	tele.Info(ctx, "GetUserGroupsPaginated called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetUserGroupsPaginated: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	offset := req.GetOffset()
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	pag := models.Pagination{
		UserId: ct.Id(userId),
		Limit:  ct.Limit(limit),
		Offset: ct.Offset(offset),
	}

	resp, err := s.Application.GetUserGroupsPaginated(ctx, pag)
	if err != nil {
		tele.Error(ctx, "Error in GetUserGroupsPaginated. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return groupsToPb(resp), nil
}

func (s *UsersHandler) GetGroupInfo(ctx context.Context, req *pb.GeneralGroupRequest) (*pb.Group, error) {
	tele.Info(ctx, "GetGroupInfo called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetGroupInfo: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	resp, err := s.Application.GetGroupInfo(ctx, models.GeneralGroupReq{
		UserId:  ct.Id(userId),
		GroupId: ct.Id(groupId),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetGroupInfo. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &pb.Group{
		GroupId:          resp.GroupId.Int64(),
		GroupOwnerId:     resp.GroupOwnerId.Int64(),
		GroupTitle:       resp.GroupTitle.String(),
		GroupDescription: resp.GroupDescription.String(),
		GroupImageId:     resp.GroupImage.Int64(),
		GroupImageUrl:    resp.GroupImageURL,
		MembersCount:     resp.MembersCount,
		IsMember:         resp.IsMember,
		IsOwner:          resp.IsOwner,
		PendingRequest:   resp.PendingRequest,
		PendingInvite:    resp.PendingInvite,
	}, nil
}

func (s *UsersHandler) GetAllGroupMemberIds(ctx context.Context, req *pb.IdReq) (*pb.Ids, error) {
	tele.Info(ctx, "GetAllGroupMemberIds called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetAllGroupMemberIds: request is nil")
	}

	memberIds, err := s.Application.GetAllGroupMemberIds(ctx, models.GroupId(req.Id))
	if err != nil {
		tele.Error(ctx, "Error in GetAllGroupMemberIds. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &pb.Ids{Ids: memberIds.Int64()}, nil
}

func (s *UsersHandler) GetGroupMembers(ctx context.Context, req *pb.GroupMembersRequest) (*pb.GroupUserArr, error) {
	tele.Info(ctx, "GetGroupMembers called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetGroupMembers: request is nil")
	}
	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	limit := req.Limit
	offset := req.Offset
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	resp, err := s.Application.GetGroupMembers(ctx, models.GroupMembersReq{
		UserId:  ct.Id(userId),
		GroupId: ct.Id(groupId),
		Limit:   ct.Limit(limit),
		Offset:  ct.Offset(offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetGroupMembers. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return groupUsersToPB(resp), nil
}

func (s *UsersHandler) GetPendingGroupJoinRequests(ctx context.Context, req *pb.GroupMembersRequest) (*cm.ListUsers, error) {
	tele.Info(ctx, "GetPendingGroupJoinRequests called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetPendingGroupJoinRequests: request is nil")
	}
	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	limit := req.Limit
	offset := req.Offset
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	resp, err := s.Application.GetPendingGroupJoinRequests(ctx, models.GroupMembersReq{
		UserId:  ct.Id(userId),
		GroupId: ct.Id(groupId),
		Limit:   ct.Limit(limit),
		Offset:  ct.Offset(offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetPendingGroupJoinRequests. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return usersToPB(resp), nil
}

func (s *UsersHandler) GetPendingGroupJoinRequestsCount(ctx context.Context, req *pb.GeneralGroupRequest) (*pb.CountResp, error) {
	tele.Info(ctx, "GetPendingGroupJoinRequestsCount called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetPendingGroupJoinRequestsCount: request is nil")
	}
	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	resp, err := s.Application.GetPendingGroupJoinRequestsCount(ctx, models.GroupJoinRequest{
		GroupId:     ct.Id(groupId),
		RequesterId: ct.Id(userId),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetPendingGroupJoinRequestsCount. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &pb.CountResp{Id: resp}, nil
}

func (s *UsersHandler) GetFollowersNotInvitedToGroup(ctx context.Context, req *pb.GroupMembersRequest) (*cm.ListUsers, error) {
	tele.Info(ctx, "GetFollowersNotInvitedToGroup called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "GetFollowersNotInvitedToGroup: request is nil")
	}
	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	limit := req.Limit
	offset := req.Offset
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	resp, err := s.Application.GetFollowersNotInvitedToGroup(ctx, models.GroupMembersReq{
		UserId:  ct.Id(userId),
		GroupId: ct.Id(groupId),
		Limit:   ct.Limit(limit),
		Offset:  ct.Offset(offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in GetFollowersNotInvitedToGroup. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return usersToPB(resp), nil
}

func (s *UsersHandler) SearchGroups(ctx context.Context, req *pb.GroupSearchRequest) (*pb.GroupArr, error) {
	tele.Info(ctx, "SearchGroups called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "SearchGroups: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}
	search := req.SearchTerm
	if err := invalidString("search", search); err != nil {
		return nil, err
	}

	limit := req.Limit
	offset := req.Offset
	if err := checkLimOff(limit, offset); err != nil {
		return nil, err
	}

	resp, err := s.Application.SearchGroups(ctx, models.GroupSearchReq{
		UserId:     ct.Id(userId),
		SearchTerm: ct.SearchTerm(search),
		Limit:      ct.Limit(limit),
		Offset:     ct.Offset(offset),
	})
	if err != nil {
		tele.Error(ctx, "Error in SearchGroups. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return groupsToPb(resp), nil
}

func (s *UsersHandler) InviteToGroup(ctx context.Context, req *pb.InviteToGroupRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "InviteToGroup called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "InviteToGroup: request is nil")
	}

	inviterId := req.InviterId
	if err := invalidId("inviterId", inviterId); err != nil {
		return nil, err
	}

	groupId := req.GroupId
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	err := s.Application.InviteToGroup(ctx, models.InviteToGroupReq{
		InviterId:  ct.Id(inviterId),
		InvitedIds: ct.FromInt64s(req.InvitedIds.Values),
		GroupId:    ct.Id(groupId),
	})
	if err != nil {
		tele.Error(ctx, "Error in InviteToGroup. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) IsGroupMember(ctx context.Context, req *pb.GeneralGroupRequest) (*wrapperspb.BoolValue, error) {
	tele.Info(ctx, "IsGroupMember called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "IsGroupMember: request is nil")
	}
	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}
	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}
	resp, err := s.Application.IsGroupMember(ctx, models.GeneralGroupReq{
		UserId:  ct.Id(userId),
		GroupId: ct.Id(groupId),
	})
	if err != nil {
		tele.Error(ctx, "Error in IsGroupMember. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return wrapperspb.Bool(resp), nil
}

func (s *UsersHandler) RequestJoinGroup(ctx context.Context, req *pb.GroupJoinRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "RequestJoinGroup called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "RequestJoinGroup: request is nil")
	}

	groupId := req.GroupId
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	requesterId := req.RequesterId
	if err := invalidId("RequesterId", requesterId); err != nil {
		return nil, err
	}

	err := s.Application.RequestJoinGroup(ctx, models.GroupJoinRequest{
		GroupId:     ct.Id(groupId),
		RequesterId: ct.Id(requesterId),
	})
	if err != nil {
		tele.Error(ctx, "Error in RequestJoinGroup. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) CancelJoinGroupRequest(ctx context.Context, req *pb.GroupJoinRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "CancelGroupJoinRequest called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "CancelGroupJoinRequest: request is nil")
	}

	groupId := req.GroupId
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	requesterId := req.RequesterId
	if err := invalidId("RequesterId", requesterId); err != nil {
		return nil, err
	}

	err := s.Application.CancelJoinGroupRequest(ctx, models.GroupJoinRequest{
		GroupId:     ct.Id(groupId),
		RequesterId: ct.Id(requesterId),
	})
	if err != nil {
		tele.Error(ctx, "Error in CancelJoinGroupRequest. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) RespondToGroupInvite(ctx context.Context, req *pb.HandleGroupInviteRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "RespondToGroupInvite called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "RespondToGroupInvite: request is nil")
	}

	groupId := req.GroupId
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	InvitedId := req.InvitedId
	if err := invalidId("InvitedId", InvitedId); err != nil {
		return nil, err
	}

	acc := req.Accepted

	err := s.Application.RespondToGroupInvite(ctx, models.HandleGroupInviteRequest{
		GroupId:   ct.Id(groupId),
		InvitedId: ct.Id(InvitedId),
		Accepted:  acc,
	})
	if err != nil {
		tele.Error(ctx, "Error in RespondToGroupInvite. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) HandleGroupJoinRequest(ctx context.Context, req *pb.HandleJoinRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "HandleGroupJoinRequest called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "HandleGroupJoinRequest: request is nil")
	}

	groupId := req.GroupId
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	RequesterId := req.RequesterId
	if err := invalidId("RequesterId", RequesterId); err != nil {
		return nil, err
	}

	ownerId := req.OwnerId
	if err := invalidId("ownerId", ownerId); err != nil {
		return nil, err
	}

	acc := req.Accepted

	err := s.Application.HandleGroupJoinRequest(ctx, models.HandleJoinRequest{
		GroupId:     ct.Id(groupId),
		RequesterId: ct.Id(RequesterId),
		OwnerId:     ct.Id(ownerId),
		Accepted:    acc,
	})
	if err != nil {
		tele.Error(ctx, "Error in HandleGroupJoinRequest. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) LeaveGroup(ctx context.Context, req *pb.GeneralGroupRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "LeaveGroup called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "LeaveGroup: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	err := s.Application.LeaveGroup(ctx, models.GeneralGroupReq{
		UserId:  ct.Id(userId),
		GroupId: ct.Id(groupId),
	})
	if err != nil {
		tele.Error(ctx, "Error in LeaveGroup. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) RemoveFromGroup(ctx context.Context, req *pb.RemoveFromGroupRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "RemoveFromGroup called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "RemoveFromGroup: request is nil")
	}

	groupId := req.GetGroupId()
	if err := invalidId("groupId", groupId); err != nil {
		return nil, err
	}

	memberId := req.GetMemberId()
	if err := invalidId("memberId", memberId); err != nil {
		return nil, err
	}

	ownerId := req.GetOwnerId()
	if err := invalidId("ownerId", ownerId); err != nil {
		return nil, err
	}

	err := s.Application.RemoveFromGroup(ctx, models.RemoveFromGroupRequest{
		GroupId:  ct.Id(groupId),
		MemberId: ct.Id(memberId),
		OwnerId:  ct.Id(ownerId),
	})
	if err != nil {
		tele.Error(ctx, "Error in RemoveFromGroup. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*wrapperspb.Int64Value, error) {
	tele.Info(ctx, "CreateGroup called @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "CreateGroup: request is nil")
	}

	OwnerId := req.OwnerId
	if err := invalidId("owner", OwnerId); err != nil {
		return nil, err
	}

	GroupTitle := req.GroupTitle
	if err := invalidString("GroupTitle", GroupTitle); err != nil {
		return nil, err
	}

	GroupDescription := req.GroupDescription
	if err := invalidString("GroupDescription", GroupDescription); err != nil {
		return nil, err
	}

	GroupImage := req.GroupImageId
	// if err := invalidId("GroupImage", GroupImage); err != nil {
	// 	return nil, err
	// }

	resp, err := s.Application.CreateGroup(ctx, &models.CreateGroupRequest{
		OwnerId:          ct.Id(OwnerId),
		GroupTitle:       ct.Title(GroupTitle),
		GroupDescription: ct.About(GroupDescription),
		GroupImage:       ct.Id(GroupImage),
	})
	if err != nil {
		tele.Error(ctx, "Error in CreateGroup. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return wrapperspb.Int64(int64(resp)), nil
}

func (s *UsersHandler) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "UpdateGroup called @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "CreateGroup: request is nil")
	}
	requesterId := req.RequesterId
	if err := invalidId("requester", requesterId); err != nil {
		return nil, err
	}

	groupId := req.GroupId
	if err := invalidId("group", groupId); err != nil {
		return nil, err
	}

	groupTitle := req.GroupTitle
	if err := invalidString("GroupTitle", groupTitle); err != nil {
		return nil, err
	}

	groupDescription := req.GroupDescription
	if err := invalidString("GroupDescription", groupDescription); err != nil {
		return nil, err
	}

	groupImage := req.GroupImageId

	err := s.Application.UpdateGroup(ctx, &models.UpdateGroupRequest{
		RequesterId:      ct.Id(requesterId),
		GroupId:          ct.Id(groupId),
		GroupTitle:       ct.Title(groupTitle),
		GroupDescription: ct.About(groupDescription),
		GroupImage:       ct.Id(groupImage),
		DeleteImage:      req.GetDeleteImage(),
	})
	if err != nil {
		tele.Error(ctx, "Error in UpdateGroup. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) GetGroupBasicInfo(ctx context.Context, req *pb.IdReq) (*pb.Group, error) {
	tele.Info(ctx, "GetGroupBasicInfo called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	g, err := s.Application.GetGroupBasicInfo(ctx, models.GroupId(req.Id))
	if err != nil {
		tele.Error(ctx, "Error in GetGroupBasicInfo. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &pb.Group{
		GroupId:          g.GroupId.Int64(),
		GroupOwnerId:     g.GroupOwnerId.Int64(),
		GroupTitle:       g.GroupTitle.String(),
		GroupDescription: g.GroupDescription.String(),
		GroupImageId:     g.GroupImage.Int64(),
	}, nil

}

// PROFILE
func (s *UsersHandler) GetBasicUserInfo(ctx context.Context, req *wrapperspb.Int64Value) (*cm.User, error) {
	tele.Info(ctx, "GetBasicUserInfo called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	userId := req.GetValue()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	u, err := s.Application.GetBasicUserInfo(ctx, ct.Id(req.GetValue()))
	if err != nil {
		tele.Error(ctx, "Error in GetBasicUserInfo. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	return &cm.User{
		UserId:    u.UserId.Int64(),
		Username:  u.Username.String(),
		Avatar:    u.AvatarId.Int64(),
		AvatarUrl: u.AvatarURL,
	}, nil
}

func (s *UsersHandler) GetBatchBasicUserInfo(ctx context.Context, req *cm.UserIds) (*cm.ListUsers, error) {
	tele.Info(ctx, "GetBatchBasicUserInfo called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	userIds := req.GetValues()
	if len(userIds) == 0 {
		return &cm.ListUsers{Users: []*cm.User{}}, nil
	}

	ids := ct.FromInt64s(userIds)
	users, err := s.Application.GetBatchBasicUserInfo(ctx, ids)
	if err != nil {
		tele.Error(ctx, "Error in GetBatchBasicUserInfo. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	pbUsers := make([]*cm.User, 0, len(users))
	for _, u := range users {
		pbUsers = append(pbUsers, &cm.User{
			UserId:    u.UserId.Int64(),
			Username:  u.Username.String(),
			Avatar:    u.AvatarId.Int64(),
			AvatarUrl: u.AvatarURL,
		})
	}

	return &cm.ListUsers{Users: pbUsers}, nil
}

func (s *UsersHandler) GetUserProfile(ctx context.Context, req *pb.GetUserProfileRequest) (*pb.UserProfileResponse, error) {
	tele.Info(ctx, "GetUserProfile gRPC method called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}
	RequesterId := req.GetRequesterId()
	if err := invalidId("RequesterId", RequesterId); err != nil {
		return nil, err
	}

	userProfileRequest := models.UserProfileRequest{
		UserId:      ct.Id(req.GetUserId()),
		RequesterId: ct.Id(req.GetRequesterId()),
	}

	profile, err := s.Application.GetUserProfile(ctx, userProfileRequest)
	if err != nil {
		tele.Error(ctx, "Error in GetUserProfile. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	tele.Debug(ctx, "get user profile @1", "profile", profile)

	return &pb.UserProfileResponse{
		UserId:                        profile.UserId.Int64(),
		Username:                      profile.Username.String(),
		FirstName:                     profile.FirstName.String(),
		LastName:                      profile.LastName.String(),
		DateOfBirth:                   profile.DateOfBirth.ToProto(),
		Avatar:                        profile.AvatarId.Int64(),
		AvatarUrl:                     profile.AvatarURL,
		About:                         profile.About.String(),
		Public:                        profile.Public,
		CreatedAt:                     profile.CreatedAt.ToProto(),
		Email:                         profile.Email.String(),
		FollowersCount:                profile.FollowersCount,
		FollowingCount:                profile.FollowingCount,
		GroupsCount:                   profile.GroupsCount,
		OwnedGroupsCount:              profile.OwnedGroupsCount,
		ViewerIsFollowing:             profile.ViewerIsFollowing,
		OwnProfile:                    profile.OwnProfile,
		IsPending:                     profile.IsPending,
		FollowRequestFromProfileOwner: profile.FollowRequestFromProfileOwner,
	}, nil
}

func (s *UsersHandler) SearchUsers(ctx context.Context, req *pb.UserSearchRequest) (*cm.ListUsers, error) {
	tele.Info(ctx, "SearchUsers called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "SearchUsers: request is nil")
	}

	SearchTerm := req.SearchTerm
	if err := invalidString("SearchTerm", SearchTerm); err != nil {
		return nil, err
	}

	limit := req.Limit
	if err := checkLimOff(limit, 1); err != nil {
		return nil, err
	}

	resp, err := s.Application.SearchUsers(ctx, models.UserSearchReq{
		SearchTerm: ct.SearchTerm(SearchTerm),
		Limit:      ct.Limit(limit),
	})
	if err != nil {
		tele.Error(ctx, "Error in SearchUsers. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return usersToPB(resp), nil
}

func (s *UsersHandler) UpdateUserProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UserProfileResponse, error) {
	tele.Info(ctx, "UpdateUserProfile called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "UpdateUserProfile: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	resp, err := s.Application.UpdateUserProfile(ctx, models.UpdateProfileRequest{
		UserId:      ct.Id(userId),
		Username:    ct.Username(req.GetUsername()),
		FirstName:   ct.Name(req.GetFirstName()),
		LastName:    ct.Name(req.GetLastName()),
		DateOfBirth: ct.DateOfBirth(req.GetDateOfBirth().AsTime()),
		AvatarId:    ct.Id(req.GetAvatar()),
		About:       ct.About(req.GetAbout()),
		DeleteImage: req.GetDeleteImage(),
	})
	if err != nil {
		tele.Error(ctx, "Error in UpdateUserProfile. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}

	dob := timestamppb.New(resp.DateOfBirth.Time())
	if resp.DateOfBirth.Time().IsZero() {
		dob = nil
	}

	return &pb.UserProfileResponse{
		UserId:      resp.UserId.Int64(),
		Username:    resp.Username.String(),
		FirstName:   resp.FirstName.String(),
		LastName:    resp.LastName.String(),
		DateOfBirth: dob,
		Avatar:      resp.AvatarId.Int64(),
		About:       resp.About.String(),
		Public:      resp.Public,
	}, nil
}

func (s *UsersHandler) UpdateProfilePrivacy(ctx context.Context, req *pb.UpdateProfilePrivacyRequest) (*emptypb.Empty, error) {
	tele.Info(ctx, "UpdateProfilePrivacy called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "UpdateProfilePrivacy: request is nil")
	}

	userId := req.GetUserId()
	if err := invalidId("userId", userId); err != nil {
		return nil, err
	}

	public := req.Public

	err := s.Application.UpdateProfilePrivacy(ctx, models.UpdateProfilePrivacyRequest{
		UserId: ct.Id(userId),
		Public: public,
	})
	if err != nil {
		tele.Error(ctx, "Error in UpdateProfilePrivacy. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *UsersHandler) RemoveImages(ctx context.Context, req *pb.FailedImageIds) (*emptypb.Empty, error) {
	tele.Info(ctx, "RemoveImages called with @1", "request", req.String())

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "RemoveImages: request is nil")
	}

	err := s.Application.RemoveImages(ctx, req.ImgIds)
	if err != nil {
		tele.Error(ctx, "Error in RemoveImages. @1", "error", err.Error(), "request", req.String())
		return nil, ce.EncodeProto(err)
	}
	return &emptypb.Empty{}, nil
}

// CONVERTORS
func usersToPB(dbUsers []models.User) *cm.ListUsers {
	pbUsers := make([]*cm.User, 0, len(dbUsers))

	for _, u := range dbUsers {
		pbUsers = append(pbUsers, &cm.User{
			UserId:    u.UserId.Int64(),
			Username:  u.Username.String(),
			Avatar:    u.AvatarId.Int64(),
			AvatarUrl: u.AvatarURL,
		})
	}

	return &cm.ListUsers{Users: pbUsers}
}

func groupsToPb(groups []models.Group) *pb.GroupArr {
	pbGroups := make([]*pb.Group, 0, len(groups))
	for _, g := range groups {
		pbGroups = append(pbGroups, &pb.Group{
			GroupId:          g.GroupId.Int64(),
			GroupOwnerId:     g.GroupOwnerId.Int64(),
			GroupTitle:       g.GroupTitle.String(),
			GroupDescription: g.GroupDescription.String(),
			GroupImageId:     g.GroupImage.Int64(),
			GroupImageUrl:    g.GroupImageURL,
			MembersCount:     g.MembersCount,
			IsMember:         g.IsMember,
			IsOwner:          g.IsOwner,
			PendingRequest:   g.PendingRequest,
			PendingInvite:    g.PendingInvite,
		})
	}

	return &pb.GroupArr{
		GroupArr: pbGroups,
	}
}

func groupUsersToPB(users []models.GroupUser) *pb.GroupUserArr {
	out := &pb.GroupUserArr{
		GroupUserArr: make([]*pb.GroupUser, 0, len(users)),
	}

	for _, u := range users {
		out.GroupUserArr = append(out.GroupUserArr, &pb.GroupUser{
			UserId:    u.UserId.Int64(),
			Username:  u.Username.String(),
			Avatar:    u.AvatarId.Int64(),
			GroupRole: u.GroupRole,
		})
	}

	return out
}

func invalidId(varName string, value int64) error {
	if value <= 0 {
		pc, _, _, ok := runtime.Caller(1)
		funcName := "unknown"
		if ok {
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				funcName = fn.Name()
			}
		}

		return status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("[%s] variable: %v, value: %v must be larger than zero", funcName, varName, value),
		)
	}
	return nil
}

func invalidString(varName string, value string) error {
	if value == "" {
		pc, _, _, ok := runtime.Caller(1)
		funcName := "unknown"
		if ok {
			fn := runtime.FuncForPC(pc)
			if fn != nil {
				funcName = fn.Name()
			}
		}

		return status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("[%s] variable: %v, value: %v must be non empty", funcName, varName, value),
		)
	}
	return nil

}

func checkLimOff(limit, offset int32) error {
	pc, _, _, ok := runtime.Caller(1)
	funcName := "unknown"
	if ok {
		fn := runtime.FuncForPC(pc)
		if fn != nil {
			funcName = fn.Name()
		}
	}

	var maxLimit int32 = 100
	if limit > maxLimit {
		return status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("[%s] limit value: %v must be less than %v", funcName, limit, maxLimit),
		)
	}

	if offset < 0 {
		return status.Error(
			codes.InvalidArgument,
			fmt.Sprintf("[%s] offset value: %v must be larger than 0", funcName, offset),
		)
	}
	return nil
}
