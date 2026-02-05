package entry

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"social-network/services/gateway/internal/handlers"
	"social-network/shared/gen-go/chat"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/notifications"
	"social-network/shared/gen-go/posts"
	"social-network/shared/gen-go/users"
	configutil "social-network/shared/go/configs"
	"social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	"social-network/shared/go/models"
	redis_connector "social-network/shared/go/redis"
	"social-network/shared/go/retrievemedia"
	"social-network/shared/go/retrieveusers"
	tele "social-network/shared/go/telemetry"
	"syscall"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

func Run() {
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	cfgs := getConfigs()

	// Inject envs to custom types
	ct.InitCustomTypes(cfgs.PassSecret, cfgs.EncrytpionKey)

	//
	//
	//
	// TELEMETRY
	closeTelemetry, err := tele.InitTelemetry(ctx, "api-gateway", "API", cfgs.TelemetryCollectorAddress, ct.CommonKeys(), cfgs.EnableDebugLogs, cfgs.SimplePrint)
	if err != nil {
		tele.Fatalf("failed to init telemetry: %s", err.Error())
	}
	defer closeTelemetry()
	tele.Info(ctx, "initialized telemetry")

	newCtx, initSpan := tele.Trace(ctx, "gateway initializing")
	ctx = newCtx
	//
	//
	//
	// CACHE
	CacheService := redis_connector.NewRedisClient(
		cfgs.SentinelAddrs,
		cfgs.RedisPassword,
		cfgs.RedisDB,
		cfgs.RedisMasterName,
	)
	if err := CacheService.TestRedisConnection(); err != nil {
		tele.Fatalf("connection test failed, ERROR: %v", err)
	}
	tele.Info(ctx, "Cache service connection started correctly")

	//
	//
	//
	// GRPC CLIENTS
	NotifService, err := gorpc.GetGRpcClient(
		notifications.NewNotificationServiceClient,
		cfgs.NotifGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to users service: %v", err)
	}

	UsersService, err := gorpc.GetGRpcClient(
		users.NewUserServiceClient,
		cfgs.UsersGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to users service: %v", err)
	}

	PostsService, err := gorpc.GetGRpcClient(
		posts.NewPostsServiceClient,
		cfgs.PostsGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to posts service: %v", err)
	}

	ChatService, err := gorpc.GetGRpcClient(
		chat.NewChatServiceClient,
		cfgs.ChatGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to chat service: %v", err)
	}

	MediaService, err := gorpc.GetGRpcClient(
		media.NewMediaServiceClient,
		cfgs.MediaGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to media service: %v", err)
	}

	localCache, err := ristretto.NewCache(&ristretto.Config[ct.Id, *models.User]{
		NumCounters: 10 * 100_000, // number of keys to track frequency of (10M).
		MaxCost:     100_000,      // maximum cost of cache (100_000 users).
		BufferItems: 64,           // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	defer localCache.Close()

	retrieveMedia := retrievemedia.NewMediaRetriever(MediaService, CacheService, 3*time.Minute)

	retriveUsers := retrieveusers.NewUserRetriever(
		UsersService,
		CacheService,
		retrieveMedia,
		3*time.Minute,
		localCache,
	)

	//
	//
	//
	// HANDLER
	apiMux := handlers.NewHandlers(
		"gateway",
		CacheService,
		UsersService,
		PostsService,
		ChatService,
		MediaService,
		NotifService,
		retriveUsers,
	)

	//
	//
	//
	// SERVER
	server := &http.Server{
		Handler:     apiMux,
		Addr:        cfgs.HTTPAddr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	srvErr := make(chan error, 1)
	go func() {
		tele.Info(ctx, "Starting server at @1", "address", server.Addr)
		initSpan.End()
		srvErr <- server.ListenAndServe()
	}()

	//
	//
	//
	// SHUTDOWN
	select {
	case err = <-srvErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			tele.Fatalf("Failed to listen and serve: %v", err)
		}
	case <-ctx.Done():
		stopSignal()
	}

	tele.Info(ctx, "Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfgs.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		tele.Fatalf("Graceful server Shutdown Failed: %v", err)
	}

	tele.Info(ctx, "Server stopped")
}

type configs struct {
	RedisAddr       string   `env:"REDIS_ADDR"`
	SentinelAddrs   []string `env:"SENTINEL_ADDRS"`
	RedisPassword   string   `env:"REDIS_PASSWORD"`
	RedisMasterName string   `env:"REDIS_MASTER"`
	RedisDB         int      `env:"REDIS_DB"`

	UsersGRPCAddr string `env:"USERS_GRPC_ADDR"`
	PostsGRPCAddr string `env:"POSTS_GRPC_ADDR"`
	ChatGRPCAddr  string `env:"CHAT_GRPC_ADDR"`
	MediaGRPCAddr string `env:"MEDIA_GRPC_ADDR"`
	NotifGRPCAddr string `env:"NOTIFICATIONS_GRPC_ADDR"`

	HTTPAddr        string `env:"HTTP_ADDR"`
	ShutdownTimeout int    `env:"SHUTDOWN_TIMEOUT_SECONDS"`

	EnableDebugLogs bool `env:"ENABLE_DEBUG_LOGS"`
	SimplePrint     bool `env:"ENABLE_SIMPLE_PRINT"`

	OtelResourceAttributes    string `env:"OTEL_RESOURCE_ATTRIBUTES"`
	TelemetryCollectorAddress string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	PassSecret                string `env:"PASSWORD_SECRET"`
	EncrytpionKey             string `env:"ENC_KEY"`
}

func getConfigs() configs { // sensible defaults
	cfgs := configs{
		RedisAddr:       "redis:6379",
		SentinelAddrs:   []string{"redis:26379"},
		RedisPassword:   "admin",
		RedisMasterName: "mymaster",
		RedisDB:         0,

		UsersGRPCAddr: "users:50051",
		PostsGRPCAddr: "posts:50051",
		ChatGRPCAddr:  "chat:50051",
		MediaGRPCAddr: "media:50051",
		NotifGRPCAddr: "notifications:50051",

		HTTPAddr:        "0.0.0.0:8081",
		ShutdownTimeout: 5,

		EnableDebugLogs:           true,
		SimplePrint:               true,
		OtelResourceAttributes:    "service.name=api-gateway,service.namespace=social-network,deployment.environment=dev",
		TelemetryCollectorAddress: "alloy:4317",
		PassSecret:                "a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=",
		EncrytpionKey:             "a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=",
	}

	// load environment variables if present
	_, err := configutil.LoadConfigs(&cfgs)
	if err != nil {
		tele.Fatalf("failed to load env variables into config struct: %v", err)
	}

	return cfgs
}
