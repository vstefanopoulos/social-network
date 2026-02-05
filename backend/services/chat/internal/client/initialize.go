package client

import (
	"context"
	cm "social-network/shared/gen-go/common"
	"social-network/shared/gen-go/media"
	userpb "social-network/shared/gen-go/users"
	rds "social-network/shared/go/redis"
)

type Clients struct {
	UserClient  userpb.UserServiceClient
	MediaClient media.MediaServiceClient
	RedisClient *rds.RedisClient
}

type RetriveUsers func(ctx context.Context, userIds *cm.UserIds) (*cm.ListUsers, error)

func NewClients(
	userClient userpb.UserServiceClient,
	mediaClient media.MediaServiceClient,
	redis *rds.RedisClient) *Clients {

	return &Clients{
		UserClient:  userClient,
		MediaClient: mediaClient,
		RedisClient: redis,
	}
}
