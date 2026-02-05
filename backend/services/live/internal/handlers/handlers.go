package handlers

import (
	"context"
	"net/http"
	"social-network/shared/gen-go/chat"
	middleware "social-network/shared/go/http-middleware"
	"social-network/shared/go/ratelimit"
	"time"

	"github.com/nats-io/nats.go"
)

type Handlers struct {
	CacheService CacheService
	ChatService  chat.ChatServiceClient
	Nats         *nats.Conn
}

type CacheService interface {
	IncrEx(ctx context.Context, key string, expSeconds int64) (int, error)
	SetStr(ctx context.Context, key string, value string, exp time.Duration) error
	GetStr(ctx context.Context, key string) (any, error)
	SetObj(ctx context.Context, key string, value any, exp time.Duration) error
	GetObj(ctx context.Context, key string, dest any) error
	Del(ctx context.Context, key string) error
}

func NewHandlers(serviceName string, CacheService CacheService, nats *nats.Conn, ChatService chat.ChatServiceClient) *http.ServeMux {
	handlers := Handlers{
		CacheService: CacheService,
		Nats:         nats,
		ChatService:  ChatService,
	}
	return handlers.BuildMux(serviceName)
}

// BuildMux builds and returns the HTTP request multiplexer with all routes and middleware applied
func (h *Handlers) BuildMux(serviceName string) *http.ServeMux {
	mux := http.NewServeMux()
	ratelimiter := ratelimit.NewRateLimiter(serviceName+":", h.CacheService)
	middlewareObj := middleware.NewMiddleware(ratelimiter, serviceName, mux)
	SetEndpoint := middlewareObj.SetEndpoint

	IP := middleware.IPLimit
	USERID := middleware.UserLimit

	SetEndpoint("/test").
		RateLimit(IP, 10, 10).
		AllowedMethod("GET").
		RateLimit(IP, 20, 5).
		EnrichContext().
		RateLimit(USERID, 10, 10).
		Finalize(h.testHandler())

	SetEndpoint("/live").
		RateLimit(IP, 10, 10).
		AllowedMethod("GET").
		Auth().
		EnrichContext().
		RateLimit(USERID, 10, 10).
		Finalize(h.Connect())

	return mux
}
