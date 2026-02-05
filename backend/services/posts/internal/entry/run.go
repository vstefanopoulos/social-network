package entry

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"social-network/services/posts/internal/application"
	"social-network/services/posts/internal/client"
	ds "social-network/services/posts/internal/db/dbservice"
	"social-network/services/posts/internal/handler"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/notifications"
	"social-network/shared/gen-go/posts"
	"social-network/shared/gen-go/users"
	configutil "social-network/shared/go/configs"
	"social-network/shared/go/ct"
	"social-network/shared/go/kafgo"
	"social-network/shared/go/models"
	rds "social-network/shared/go/redis"
	tele "social-network/shared/go/telemetry"

	"social-network/shared/go/gorpc"
	postgresql "social-network/shared/go/postgre"
	"syscall"

	"github.com/dgraph-io/ristretto/v2"
)

func Run() error {
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	cfgs := getConfigs()

	//
	//
	//
	// TELEMETRY
	closeTelemetry, err := tele.InitTelemetry(ctx, "posts", "PST", cfgs.TelemetryCollectorAddress, ct.CommonKeys(), cfgs.EnableDebugLogs, cfgs.SimplePrint)
	if err != nil {
		tele.Fatalf("failed to init telemetry: %s", err.Error())
	}
	defer closeTelemetry()

	tele.Info(ctx, "initialized telemetry")

	//
	//
	//
	// DATABASE
	dbUrl := os.Getenv("DATABASE_URL")
	pool, err := postgresql.NewPool(ctx, dbUrl)
	if err != nil {
		return fmt.Errorf("failed to connect db: %v", err)
	}
	defer pool.Close()
	tele.Info(ctx, "Connected to posts-db database")

	//
	//
	//
	// GRPC CLIENTS
	UsersService, err := gorpc.GetGRpcClient(
		users.NewUserServiceClient,
		cfgs.UsersGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to users service: %v", err)
	}

	MediaService, err := gorpc.GetGRpcClient(
		media.NewMediaServiceClient,
		cfgs.MediaGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to media service: %v", err)
	}

	NotifService, err := gorpc.GetGRpcClient(
		notifications.NewNotificationServiceClient,
		cfgs.NotifGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to notifications service: %v", err)
	}
	if NotifService == nil {
		tele.Fatal("NotifService is nil after initialization")
	}

	//
	//
	//
	// KAFKA PRODUCER
	eventProducer, close, err := kafgo.NewKafkaProducer([]string{cfgs.KafkaBrokers})
	if err != nil {
		tele.Warn(ctx, "failed to initialize kafka producer: @1", "error", err.Error())
	} else {
		defer close()
		tele.Info(ctx, "initialized kafka producer")
	}

	//
	//
	//
	// REDIS
	redisConnector := rds.NewRedisClient(cfgs.SentinelAddrs, cfgs.RedisPassword, cfgs.RedisDB, cfgs.RedisMasterName)
	if err := redisConnector.TestRedisConnection(); err != nil {
		tele.Fatalf("connection test failed, ERROR: %v", err)
	}

	clients := client.NewClients(UsersService, MediaService, NotifService)

	//
	//
	//
	// Local Cache
	localCache, err := ristretto.NewCache(&ristretto.Config[ct.Id, *models.User]{
		NumCounters: 10 * 100_000, // number of keys to track frequency of (10M).
		MaxCost:     100_000,      // maximum cost of cache (100_000 users).
		BufferItems: 64,           // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	defer localCache.Close()

	app, err := application.NewApplication(ds.New(pool), pool, clients, redisConnector, eventProducer, localCache)
	if err != nil {
		return fmt.Errorf("failed to create posts application: %v", err)
	}

	service := handler.NewPostsHandler(app)
	tele.Info(ctx, "Running gRpc service...")

	//
	//
	//
	// GRPC SERVER
	startServerFunc, endServerFunc, err := gorpc.CreateGRpcServer[posts.PostsServiceServer](
		posts.RegisterPostsServiceServer,
		service,
		cfgs.GrpcServerPort,
		ct.CommonKeys())
	if err != nil {
		return err
	}
	defer endServerFunc()

	go func() {
		err := startServerFunc()
		if err != nil {
			tele.Fatal("server failed to start")
		}
		tele.Info(ctx, "server finished")
	}()

	//
	//
	//
	// SHUTDOWN
	// wait here for process termination signal to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	tele.Info(ctx, "Shutting down server...")
	endServerFunc()
	tele.Info(ctx, "Server stopped")
	return nil

}

type configs struct {
	RedisAddr       string   `env:"REDIS_ADDR"`
	SentinelAddrs   []string `env:"SENTINEL_ADDRS"`
	RedisPassword   string   `env:"REDIS_PASSWORD"`
	RedisMasterName string   `env:"REDIS_MASTER"`
	RedisDB         int      `env:"REDIS_DB"`

	UsersGRPCAddr  string `env:"USERS_GRPC_ADDR"`
	PostsGRPCAddr  string `env:"POSTS_GRPC_ADDR"`
	ChatGRPCAddr   string `env:"CHAT_GRPC_ADDR"`
	MediaGRPCAddr  string `env:"MEDIA_GRPC_ADDR"`
	NotifGRPCAddr  string `env:"NOTIFICATIONS_GRPC_ADDR"`
	GrpcServerPort string `env:"GRPC_SERVER_PORT"`

	KafkaBrokers string `env:"KAFKA_BROKERS"`

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
		GrpcServerPort:  ":50051",

		KafkaBrokers: "kafka:9092",

		EnableDebugLogs:           true,
		SimplePrint:               true,
		OtelResourceAttributes:    "service.name=posts,service.namespace=social-network,deployment.environment=dev",
		TelemetryCollectorAddress: "alloy:4317",

		PassSecret:    "a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=",
		EncrytpionKey: "a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=",
	}

	// load environment variables if present
	_, err := configutil.LoadConfigs(&cfgs)
	if err != nil {
		tele.Fatalf("failed to load env variables into config struct: %v", err)
	}

	return cfgs
}
