package retrieveusers

import (
	"context"
	cm "social-network/shared/gen-go/common"
	userpb "social-network/shared/gen-go/users"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	redis_connector "social-network/shared/go/redis"
	"social-network/shared/go/retrievemedia"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type UserInfoRetriever interface {
	GetBasicUserInfo(ctx context.Context, req *wrapperspb.Int64Value, opts ...grpc.CallOption) (*cm.User, error)
	GetBatchBasicUserInfo(ctx context.Context, req *cm.UserIds, opts ...grpc.CallOption) (*cm.ListUsers, error)
	RemoveImages(ctx context.Context, req *userpb.FailedImageIds, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type UserRetriever struct {
	client         UserInfoRetriever
	cache          RedisCache
	mediaRetriever *retrievemedia.MediaRetriever
	ttl            time.Duration
	localCache     *ristretto.Cache[ct.Id, *models.User]
}

// UserRetriever provides a function that abstracts the process of populating a map[ct.Id]models.User
// from a slice of user ids. It depends on:
//
//  1. GetBatchBasicUserInfo, GetBasicUserInfo, RemoveImage functions provided by an initiator that holds a connection to social-network user service.
//  2. A cache interface that implements GetObj() and SetObj() methods.
//  3. The retrievemedia package.
func NewUserRetriever(client UserInfoRetriever, cache *redis_connector.RedisClient, mediaRetriever *retrievemedia.MediaRetriever, ttl time.Duration, localCache *ristretto.Cache[ct.Id, *models.User]) *UserRetriever {
	return &UserRetriever{client: client, cache: cache, mediaRetriever: mediaRetriever, ttl: ttl, localCache: localCache}
}

// RedisCache defines the minimal Redis operations used by the hydrator.
type RedisCache interface {
	GetObj(ctx context.Context, key string, dest any) error
	SetObj(ctx context.Context, key string, value any, exp time.Duration) error
}
