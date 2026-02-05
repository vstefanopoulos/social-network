package client

import (
	"context"
	"fmt"
	cm "social-network/shared/gen-go/common"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/users"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
)

// Function to inject to retrive users
func (c *Clients) GetBatchBasicUserInfo(ctx context.Context, req *cm.UserIds) (*cm.ListUsers, *ce.Error) {
	resp, err := c.UserClient.GetBatchBasicUserInfo(ctx, req)
	if err != nil {
		return nil, ce.DecodeProto(err, req.String())
	}
	return resp, nil
}

// function to inject to retrieve media
func (c *Clients) GetImages(ctx context.Context, req *media.GetImagesRequest) (*media.GetImagesResponse, *ce.Error) {
	resp, err := c.MediaClient.GetImages(ctx, req)
	if err != nil {
		return nil, ce.DecodeProto(err, req.String())
	}
	return resp, nil
}

func (c *Clients) IsGroupMember(ctx context.Context,
	groupId ct.Id, userId ct.Id) (bool, *ce.Error) {
	input := fmt.Sprintf("groupId: %v, userId: %v", groupId, userId)
	resp, err := c.UserClient.IsGroupMember(ctx, &users.GeneralGroupRequest{
		GroupId: groupId.Int64(),
		UserId:  userId.Int64(),
	})
	if err != nil {
		return false, ce.DecodeProto(err, input)
	}
	return resp.GetValue(), nil
}

func (c *Clients) AreConnected(ctx context.Context, userA, userB ct.Id) (bool, *ce.Error) {
	input := fmt.Sprintf("userA: %v, userB: %v", userA, userB)
	resp, err := c.UserClient.AreFollowingEachOther(ctx, &users.FollowUserRequest{
		FollowerId:   userA.Int64(),
		TargetUserId: userB.Int64(),
	})
	if err != nil {
		return false, ce.DecodeProto(err, input)
	}

	connected := resp.FollowerFollowsTarget || resp.TargetFollowsFollower
	return connected, nil
}
