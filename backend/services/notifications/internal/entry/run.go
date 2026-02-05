package entry

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"social-network/services/notifications/internal/application"
	"social-network/services/notifications/internal/client"
	"social-network/services/notifications/internal/db/sqlc"
	"social-network/services/notifications/internal/events"
	"social-network/services/notifications/internal/handler"
	"social-network/shared/gen-go/chat"
	pb "social-network/shared/gen-go/notifications"
	"social-network/shared/gen-go/posts"
	"social-network/shared/gen-go/users"
	configutil "social-network/shared/go/configs"
	"social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"

	"social-network/shared/go/gorpc"
	"social-network/shared/go/kafgo"
	postgresql "social-network/shared/go/postgre"
	"syscall"

	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func Run() error {
	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal() //check if ok

	cfgs := getConfigs()

	pool, err := postgresql.NewPool(ctx, cfgs.DatabaseConn)
	if err != nil {
		return fmt.Errorf("failed to connect db: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to notifications database")

	postsService, err := gorpc.GetGRpcClient(
		posts.NewPostsServiceClient,
		cfgs.PostsGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		log.Fatalf("failed to connect to posts service: %v", err)
	}

	chatService, err := gorpc.GetGRpcClient(
		chat.NewChatServiceClient,
		cfgs.ChatGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		log.Fatalf("failed to connect to chat service: %v", err)
	}

	usersService, err := gorpc.GetGRpcClient(
		users.NewUserServiceClient,
		cfgs.UsersGRPCAddr,
		ct.CommonKeys(),
	)
	if err != nil {
		log.Fatalf("failed to connect to media service: %v", err)
	}

	// Initialize NATS connection
	natsConn, err := nats.Connect(cfgs.NatsCluster)
	if err != nil {
		log.Fatalf("failed to connect to nats: %v", err)
	}
	defer natsConn.Drain()
	log.Println("NATS connected")

	clients := client.NewClients(usersService, chatService, postsService)
	app := application.NewApplication(sqlc.New(pool), clients, natsConn)

	// Initialize default notification types
	if err := app.CreateDefaultNotificationTypes(context.Background()); err != nil {
		log.Printf("Warning: failed to create default notification types: %v", err)
	}

	// Initialize Kafka consumer and start processing
	if err := startKafkaConsumer(ctx, app); err != nil {
		return fmt.Errorf("failed to start kafka consumer: %w", err)
	}

	service := handler.NewNotificationsHandler(app)

	log.Println("Running gRpc service...")
	startServerFunc, endServerFunc, err := gorpc.CreateGRpcServer[pb.NotificationServiceServer](
		pb.RegisterNotificationServiceServer,
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
			log.Fatal("server failed to start")
		}
		fmt.Println("server finished")
	}()

	// wait here for process termination signal to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server...")
	endServerFunc()
	log.Println("Server stopped")
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
	GrpcServerPort string `env:"GRPC_SERVER_PORT"`

	KafkaBrokers []string `env:"KAFKA_BROKERS"` // Comma-separated list of Kafka brokers
	NatsCluster  string   `env:"NATS_CLUSTER"`  // NATS cluster connection string

	DatabaseConn string `env:"DATABASE_URL"`
}

func getConfigs() configs { // sensible defaults
	cfgs := configs{
		RedisAddr:       "redis:6379",
		SentinelAddrs:   []string{"redis:26379"},
		RedisPassword:   "admin",
		RedisMasterName: "mymaster",
		RedisDB:         0,
		UsersGRPCAddr:   "users:50051",
		PostsGRPCAddr:   "posts:50051",
		ChatGRPCAddr:    "chat:50051",
		KafkaBrokers:    []string{"kafka:9092"},                                                  // Default Kafka broker
		NatsCluster:     "nats://ruser:T0pS3cr3t@nats-1:4222,nats://ruser:T0pS3cr3t@nats-2:4222", // Default NATS cluster
	}

	// load environment variables if present
	_, err := configutil.LoadConfigs(&cfgs)
	if err != nil {
		tele.Fatalf("failed to load env variables into config struct: %v", err)
	}
	fmt.Println("brokers:", cfgs.KafkaBrokers)
	return cfgs
}

// startKafkaConsumer initializes and starts the Kafka consumer
func startKafkaConsumer(ctx context.Context, app *application.Application) error {
	cfgs := getConfigs()

	if len(cfgs.KafkaBrokers) == 0 {
		cfgs.KafkaBrokers = []string{"kafka:9092"} // Default broker
	}
	fmt.Println("borkers2:", cfgs.KafkaBrokers)
	tele.Info(context.Background(), fmt.Sprintln("borkers2:", cfgs.KafkaBrokers))
	// Initialize Kafka consumer
	kafkaConsumer, err := kafgo.NewKafkaConsumer(
		cfgs.KafkaBrokers,
		"notifications", // Consumer group name for notifications
		ct.NotificationTopic,
	)
	if err != nil {
		tele.Error(ctx, "failed to create kafka consumer: @1", "error", err.Error())
		return fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	// Configure buffer sizes
	kafkaConsumer = kafkaConsumer.WithCommitBuffer(100)

	// Initialize event handler
	eventHandler := events.NewEventHandler(app)

	// Start the Kafka consumer
	notificationChannel, closeConsumer, err := kafkaConsumer.StartConsuming(ctx)
	if err != nil {
		tele.Error(ctx, "failed to start kafka consumer: @1", "error", err.Error())
		return fmt.Errorf("failed to start kafka consumer: %w", err)
	}

	// Start a goroutine to handle incoming notification events
	go func() {
		defer closeConsumer()
		for {
			select {
			case <-ctx.Done():
				tele.Info(ctx, "kafka listener context done")
				return
			case record, ok := <-notificationChannel:
				if !ok {
					tele.Info(ctx, "Notification channel closed")
					return
				}
				tele.Info(ctx, "Received message from kafka")

				// Process the incoming protobuf notification event
				if err := processNotificationEvent(ctx, record, eventHandler); err != nil {
					tele.Error(ctx, "Failed to process notification event", "error", err.Error())
					// Don't commit the record if processing failed
					continue
				}

				// Commit the record after successful processing
				if err := record.Commit(ctx); err != nil {
					tele.Error(ctx, "Failed to commit notification record", "error", err)
				}
			}
		}
	}()

	return nil
}

// processNotificationEvent processes a single notification event from Kafka
func processNotificationEvent(ctx context.Context, record *kafgo.Record, eventHandler *events.EventHandler) error {
	// Deserialize the NotificationEvent wrapper
	var notificationEvent pb.NotificationEvent
	if err := proto.Unmarshal(record.Data(ctx), &notificationEvent); err != nil {
		return fmt.Errorf("failed to unmarshal protobuf notification event: %w", err)
	}

	// Process the event using the event handler
	return eventHandler.Handle(ctx, &notificationEvent)
}
