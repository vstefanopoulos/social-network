package entry

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"social-network/services/users/internal/application"
	"social-network/services/users/internal/client"
	ds "social-network/services/users/internal/db/dbservice"
	"social-network/services/users/internal/handler"
	"social-network/shared/gen-go/chat"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/notifications"
	"social-network/shared/gen-go/users"
	configutil "social-network/shared/go/configs"
	"social-network/shared/go/ct"
	"social-network/shared/go/kafgo"
	rds "social-network/shared/go/redis"
	tele "social-network/shared/go/telemetry"

	"social-network/shared/go/gorpc"
	postgresql "social-network/shared/go/postgre"
	"syscall"
)

//TODO add logs as things are getting initialized

func Run() error {
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal() //TODO REMOVE, stop signal should be called when appropriat, during shutdown of server or somethig

	cfgs := getConfigs()

	//
	//
	//
	// TELEMETRY
	closeTelemetry, err := tele.InitTelemetry(ctx, "users", "USR", cfgs.TelemetryCollectorAddress, ct.CommonKeys(), cfgs.EnableDebugLogs, cfgs.SimplePrint)
	if err != nil {
		tele.Fatalf("failed to init telemetry: %s", err.Error())
	}
	defer closeTelemetry()
	tele.Info(ctx, "initialized telemetry")

	//
	//
	//
	// CLIENT SERVICES
	chatClient, err := gorpc.GetGRpcClient(
		chat.NewChatServiceClient,
		cfgs.ChatGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatal("failed to create chat client")
	}
	mediaClient, err := gorpc.GetGRpcClient(
		media.NewMediaServiceClient,
		cfgs.MediaGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatal("failed to create media client")
	}
	notificationsClient, err := gorpc.GetGRpcClient(
		notifications.NewNotificationServiceClient,
		cfgs.NotificationsGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatal("failed to create chat client")
	}

	redisConnector := rds.NewRedisClient(cfgs.SentinelAddrs, cfgs.RedisPassword, cfgs.RedisDB, cfgs.RedisMasterName)

	//
	//
	// DATABASE
	pool, err := postgresql.NewPool(ctx, cfgs.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect db: %v", err)
	}
	defer pool.Close()

	pgxTxRunner, err := postgresql.NewPgxTxRunner(pool, ds.New(pool))
	if err != nil {
		tele.Fatal("failed to create pgxTxRunner")
	}
	tele.Info(ctx, "Connected to users-db database")

	//
	//
	//
	// KAFKA PRODUCER
	eventProducer, close, err := kafgo.NewKafkaProducer([]string{cfgs.KafkaBrokers})
	if err != nil {
		tele.Warn(ctx, "failed to initialize kafka producer: @1", "error", err.Error())
		defer close()
		tele.Info(ctx, "initialized kafka producer")
	}

	//
	//
	// APPLICATION
	clients := client.NewClients(
		chatClient,
		notificationsClient,
		mediaClient,
		redisConnector,
	)

	app := application.NewApplication(ds.New(pool), pgxTxRunner, pool, clients, eventProducer)
	service := *handler.NewUsersHanlder(app)

	//
	//
	//
	// SERVER
	startServerFunc, stopServerFunc, err := gorpc.CreateGRpcServer[users.UserServiceServer](
		users.RegisterUserServiceServer,
		&service,
		cfgs.GrpcServerPort,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("couldn't create gRpc Server: %s", err.Error())
	}

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
	tele.Info(ctx, "gRPC server listening on @1", "port", cfgs.GrpcServerPort)

	// wait here for process termination signal to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	tele.Info(ctx, "Shutting down server...")
	stopServerFunc()
	tele.Info(ctx, "Server stopped")
	return nil
}

type configs struct {
	RedisAddr       string   `env:"REDIS_ADDR"`
	SentinelAddrs   []string `env:"SENTINEL_ADDRS"`
	RedisPassword   string   `env:"REDIS_PASSWORD"`
	RedisMasterName string   `env:"REDIS_MASTER"`
	RedisDB         int      `env:"REDIS_DB"`

	DatabaseURL           string `env:"DATABASE_URL"`
	ChatGRPCAddr          string `env:"CHAT_GRPC_ADDR"`
	MediaGRPCAddr         string `env:"MEDIA_GRPC_ADDR"`
	NotificationsGRPCAddr string `env:"NOTIFICATIONS_GRPC_ADDR"`
	ShutdownTimeout       int    `env:"SHUTDOWN_TIMEOUT_SECONDS"`
	GrpcServerPort        string `env:"GRPC_SERVER_PORT"`

	KafkaBrokers string `env:"KAFKA_BROKERS"`

	OtelResourceAttributes    string `end:"OTEL_RESOURCE_ATTRIBUTES"`
	TelemetryCollectorAddress string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	EnableDebugLogs           bool   `env:"ENABLE_DEBUG_LOGS"`
	SimplePrint               bool   `env:"ENABLE_SIMPLE_PRINT"`
}

func getConfigs() configs {
	cfgs := configs{
		RedisAddr:                 "redis:6379",
		SentinelAddrs:             []string{"redis:26379"},
		RedisPassword:             "admin",
		RedisMasterName:           "mymaster",
		RedisDB:                   0,
		DatabaseURL:               "postgres://postgres:secret@users-db:5432/social_users?sslmode=disable",
		ChatGRPCAddr:              "chat:50051",
		MediaGRPCAddr:             "media:50051",
		NotificationsGRPCAddr:     "notifications:50051",
		ShutdownTimeout:           5,
		KafkaBrokers:              "kafka:9092",
		OtelResourceAttributes:    "service.name=users,service.namespace=social-network,deployment.environment=dev",
		TelemetryCollectorAddress: "alloy:4317",
		EnableDebugLogs:           true,
		SimplePrint:               true,
	}

	_, err := configutil.LoadConfigs(&cfgs)
	if err != nil {
		tele.Fatalf("failed to load env variables into config struct: %v", err)
	}

	return cfgs
}
