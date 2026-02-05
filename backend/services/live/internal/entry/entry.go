package entry

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"

	"social-network/services/live/internal/handlers"
	"social-network/shared/gen-go/chat"
	configutil "social-network/shared/go/configs"
	"social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	redis_connector "social-network/shared/go/redis"
	tele "social-network/shared/go/telemetry"
	testnats "social-network/shared/go/test-nats"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
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
	closeTelemetry, err := tele.InitTelemetry(ctx, "live", "LIV", cfgs.TelemetryCollectorAddress, ct.CommonKeys(), cfgs.EnableDebugLogs, cfgs.SimplePrint)
	if err != nil {
		tele.Fatalf("failed to init telemetry: %s", err.Error())
	}
	defer closeTelemetry()
	tele.Info(ctx, "initialized telemetry")

	//
	//
	//
	// CACHE
	cacheService := redis_connector.NewRedisClient(
		cfgs.SentinelAddrs,
		cfgs.RedisPassword,
		cfgs.RedisDB,
		cfgs.RedisMasterName,
	)
	if err := cacheService.TestRedisConnection(); err != nil {
		tele.Fatalf("connection test failed, ERROR: %v", err)
	}
	tele.Info(ctx, "Cache service connection started correctly")

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

	err = testnats.TestNATS(natsConn)
	if err != nil {
		tele.Error(ctx, "nats connection test failed: %v", err)
	}

	//
	//
	//
	// GRPC SERVICES
	ChatService, err := gorpc.GetGRpcClient(
		chat.NewChatServiceClient,
		cfgs.ChatGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		tele.Fatalf("failed to connect to chat service: %v", err)
	}

	//
	//
	//
	// HANDLER
	liveMux := handlers.NewHandlers(
		"live",
		cacheService,
		natsConn,
		ChatService,
	)

	//
	//
	//
	// SERVER
	server := &http.Server{
		Handler:     liveMux,
		Addr:        cfgs.HTTPAddr,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	srvErr := make(chan error, 1)
	go func() {
		tele.Info(ctx, "Starting server at @1", "address", server.Addr)
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

	HTTPAddr        string `env:"HTTP_ADDR"`
	ShutdownTimeout int    `env:"SHUTDOWN_TIMEOUT_SECONDS"`

	EnableDebugLogs bool `env:"ENABLE_DEBUG_LOGS"`
	SimplePrint     bool `env:"ENABLE_SIMPLE_PRINT"`

	OtelResourceAttributes    string `env:"OTEL_RESOURCE_ATTRIBUTES"`
	TelemetryCollectorAddress string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	PassSecret                string `env:"PASSWORD_SECRET"`
	EncrytpionKey             string `env:"ENC_KEY"`

	NatsHost    string `env:"NATS_HOST"`
	NatsCluster string `env:"NATS_CLUSTER"`

	ChatGRPCAddr string `env:"CHAT_GRPC_ADDR"`
}

func getConfigs() configs { // sensible defaults
	cfgs := configs{
		RedisAddr:                 "redis:6379",
		SentinelAddrs:             []string{"redis:26379"},
		RedisPassword:             "admin",
		RedisMasterName:           "mymaster",
		RedisDB:                   0,
		HTTPAddr:                  "0.0.0.0:8082",
		ShutdownTimeout:           5,
		EnableDebugLogs:           true,
		SimplePrint:               true,
		OtelResourceAttributes:    "service.name=live,service.namespace=social-network,deployment.environment=dev",
		TelemetryCollectorAddress: "alloy:4317",
		PassSecret:                "a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=",
		EncrytpionKey:             "a2F0LWFsZXgtdmFnLXlwYXQtc3RhbS16b25lMDEtZ28=",
		NatsHost:                  "nats",
		NatsCluster:               "NATS_CLUSTER",
		ChatGRPCAddr:              "chat:50051",
	}

	// load environment variables if present
	_, err := configutil.LoadConfigs(&cfgs)
	if err != nil {
		tele.Fatalf("failed to load env variables into config struct: %v", err)
	}
	return cfgs
}
