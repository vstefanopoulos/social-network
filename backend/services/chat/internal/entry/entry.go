package entry

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"social-network/services/chat/internal/application"
	"social-network/services/chat/internal/client"
	"social-network/services/chat/internal/db/dbservice"
	"social-network/services/chat/internal/handler"
	"social-network/shared/gen-go/chat"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/users"
	configutil "social-network/shared/go/configs"
	"social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	"social-network/shared/go/models"
	postgresql "social-network/shared/go/postgre"
	rds "social-network/shared/go/redis"
	"social-network/shared/go/retrievemedia"
	"social-network/shared/go/retrieveusers"
	tele "social-network/shared/go/telemetry"
	"time"

	"syscall"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/nats-io/nats.go"
)

type configs struct {
	RedisAddr                 string   `env:"REDIS_ADDR"`
	SentinelAddrs             []string `env:"SENTINEL_ADDRS"`
	RedisPassword             string   `env:"REDIS_PASSWORD"`
	RedisDB                   int      `env:"REDIS_DB"`
	RedisMasterName           string   `env:"REDIS_MASTER"`
	DatabaseConn              string   `env:"DATABASE_URL"`
	GrpcServerPort            string   `env:"GRPC_SERVER_PORT"`
	UsersAdress               string   `env:"USERS_GRPC_ADDR"`
	MediaGRPCAddr             string   `env:"MEDIA_GRPC_ADDR"`
	EnableDebugLogs           bool     `env:"ENABLE_DEBUG_LOGS"`
	SimplePrint               bool     `env:"ENABLE_SIMPLE_PRINT"`
	OtelResourceAttributes    string   `end:"OTEL_RESOURCE_ATTRIBUTES"`
	TelemetryCollectorAddress string   `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	NatsHost                  string   `env:"NATS_HOST"`
	NatsCluster               string   `env:"NATS_CLUSTER"`
	KafkaBrokers              []string `end:"KAFKA_BROKERS"`
}

var cfgs configs

// TODO add missing default values --V
func init() {
	cfgs = configs{
		DatabaseConn:    "postgres://postgres:secret@chat-db:5432/social_chat?sslmode=disable",
		GrpcServerPort:  ":50051",
		UsersAdress:     "users:50051",
		NatsHost:        "nats",
		RedisPassword:   "admin",
		RedisMasterName: "mymaster",
		NatsCluster:     "nats://ruser:T0pS3cr3t@nats-1:4222,nats://ruser:T0pS3cr3t@nats-2:4222",
		KafkaBrokers:    []string{"kafka:9092"},
		SentinelAddrs:   []string{"26379"},
	}
	configutil.LoadConfigs(&cfgs)
}

func Run() error {
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	//
	//
	//
	// TELEMETRY
	closeTele, err := tele.InitTelemetry(ctx,
		"chat",
		"CHA",
		cfgs.TelemetryCollectorAddress,
		ct.CommonKeys(),
		cfgs.EnableDebugLogs,
		cfgs.SimplePrint,
	)
	if err != nil {
		tele.Fatalf("failed to init telemetry: %s", err.Error())
	}
	defer closeTele()

	tele.Info(ctx, "initialized telemetry")

	//
	//
	//
	// DATABASE
	pool, err := postgresql.NewPool(ctx, cfgs.DatabaseConn)
	if err != nil {
		return fmt.Errorf("failed to connect db: %v", err)
	}
	defer pool.Close()
	tele.Info(ctx, "Connected to DB")

	//
	//
	//
	// GRPC SERVICES
	clients := initClients()

	localCache, err := ristretto.NewCache(&ristretto.Config[ct.Id, *models.User]{
		NumCounters: 10 * 100_000, // number of keys to track frequency of (10M).
		MaxCost:     100_000,      // maximum cost of cache (100_000 users).
		BufferItems: 64,           // number of keys per Get buffer.
	})
	if err != nil {
		panic(err)
	}
	defer localCache.Close()

	retrieveMedia := retrievemedia.NewMediaRetriever(
		clients.MediaClient,
		clients.RedisClient,
		3*time.Minute,
	)

	retriveUsers := retrieveusers.NewUserRetriever(
		clients.UserClient,
		clients.RedisClient,
		retrieveMedia,
		3*time.Minute,
		localCache,
	)

	//
	//
	//
	// NATS
	natsConn, err := nats.Connect(cfgs.NatsCluster)
	if err != nil {
		tele.Fatalf("failed to connect to nats: %s", err.Error())
	}
	defer natsConn.Drain()
	tele.Info(ctx, "NATS connected")

	//
	//
	//
	// CORE SERVICE
	app, err := application.NewChatService(
		pool,
		clients,
		dbservice.New(pool),
		retriveUsers,
		natsConn,
	)
	if err != nil {
		tele.Fatalf("failed to create chat service application: %s", err.Error())
	}

	handler := handler.NewChatHandler(app)

	//
	//
	//
	// SERVER
	startServerFunc, stopServerFunc, err := gorpc.CreateGRpcServer[chat.ChatServiceServer](
		chat.RegisterChatServiceServer,
		handler,
		cfgs.GrpcServerPort,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to create server: %s", err.Error())
	}

	go func() {
		tele.Info(ctx, "Starting grpc server at port: @1", "port", cfgs.GrpcServerPort)
		err := startServerFunc()
		if err != nil {
			tele.Fatal("server failed to start")
		}
		tele.Info(ctx, "server finished")
	}()

	tele.Info(ctx, "gRPC server listening on @1", "port", cfgs.GrpcServerPort)

	//
	//
	//
	// SHUTDOWN: wait here for process termination signal to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	tele.Info(ctx, "Shutting down server...")
	stopServerFunc()
	tele.Info(ctx, "Server stopped")
	return nil
}

func initClients() *client.Clients {
	//
	//
	//
	// GRPC SERVICES
	userClient, err := gorpc.GetGRpcClient(
		users.NewUserServiceClient, cfgs.UsersAdress, ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to create user client: %s", err.Error())
	}

	mediaClient, err := gorpc.GetGRpcClient(
		media.NewMediaServiceClient,
		cfgs.MediaGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to create media client: %s", err.Error())
	}

	//
	//
	//
	// CACHE
	redisClient := rds.NewRedisClient(
		cfgs.SentinelAddrs, cfgs.RedisPassword, cfgs.RedisDB, cfgs.RedisMasterName,
	)
	if err := redisClient.TestRedisConnection(); err != nil {
		tele.Fatalf("connection test failed, ERROR: %v", err)
	}

	return client.NewClients(
		userClient,
		mediaClient,
		redisClient,
	)
}
