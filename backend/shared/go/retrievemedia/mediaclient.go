package retrievemedia

import (
	"context"
	"social-network/shared/gen-go/media"
	"time"

	"google.golang.org/grpc"
)

type Client interface {
	GetImages(ctx context.Context, req *media.GetImagesRequest, variant *media.FileVariant) (*media.GetImagesResponse, error)
	GetImage(ctx context.Context, req *media.GetImageRequest) (*media.GetImageResponse, error)
}

// RedisCache defines the minimal Redis operations used by the hydrator.
type RedisCache interface {
	GetStr(ctx context.Context, key string) (any, error)
	SetStr(ctx context.Context, key string, value string, exp time.Duration) error
}

// MediaInfoRetriever defines the interface for fetching media info (usually gRPC client).
type MediaInfoRetriever interface {
	GetImages(ctx context.Context, in *media.GetImagesRequest, opts ...grpc.CallOption) (*media.GetImagesResponse, error)
	GetImage(ctx context.Context, in *media.GetImageRequest, opts ...grpc.CallOption) (*media.GetImageResponse, error)
}

type MediaRetriever struct {
	client MediaInfoRetriever
	cache  RedisCache
	ttl    time.Duration
}

func NewMediaRetriever(client MediaInfoRetriever, cache RedisCache, ttl time.Duration) *MediaRetriever {
	return &MediaRetriever{client: client, cache: cache, ttl: ttl}
}
