package entry

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"social-network/services/media/internal/application"
	"social-network/services/media/internal/client"
	"social-network/services/media/internal/configs"
	"social-network/services/media/internal/convertor"
	"social-network/services/media/internal/db/dbservice"
	"social-network/services/media/internal/handler"
	"social-network/services/media/internal/validator"
	"social-network/shared/gen-go/media"
	"social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	postgresql "social-network/shared/go/postgre"
	tele "social-network/shared/go/telemetry"

	"syscall"

	"github.com/minio/minio-go/v7"
)

func Run() error {
	cfgs := getConfigs()

	ctx, stopSignal := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignal()

	// Todo: Check ctx functionality on shut down
	closeTele, err := tele.InitTelemetry(ctx,
		"media",
		"MED",
		cfgs.Tele.TelemetryCollectorAddress,
		ct.CommonKeys(),
		cfgs.Tele.EnableDebugLogs,
		cfgs.Tele.SimplePrint,
	)
	if err != nil {
		tele.Fatalf("failed to init telemetry: %s", err.Error())
	}
	defer closeTele()

	tele.Info(ctx, "initialized telemetry")

	pool, err := postgresql.NewPool(ctx, cfgs.DB.URL)
	if err != nil {
		return fmt.Errorf("failed to connect db: %v", err)
	}
	defer pool.Close()

	tele.Info(ctx, "Connected to media database")

	// Internal client for backend operations
	fileServiceClient, err := NewMinIOConn(ctx, cfgs.FileService, cfgs.FileService.Endpoint, false)
	if err != nil {
		return err
	}

	// Optional public client for URL generation (e.g. localhost in dev)
	var publicFileServiceClient *minio.Client
	if cfgs.FileService.PublicEndpoint != "" {
		publicFileServiceClient, err = NewMinIOConn(ctx, cfgs.FileService, cfgs.FileService.PublicEndpoint, true)
		if err != nil {
			tele.Info(ctx, "Warning: failed to initialize public MinIO client: @1", "error", err)
		} else {
			tele.Info(ctx, "Initialized public MinIO client for URL generation")
		}
	}

	querier := dbservice.NewQuerier(pool)
	app, err := application.NewMediaService(
		pool,
		&client.Clients{
			Configs:           cfgs.FileService,
			MinIOClient:       fileServiceClient,
			PublicMinIOClient: publicFileServiceClient,
			Validator: &validator.ImageValidator{
				Config: cfgs.FileService.FileConstraints,
			},
			ImageConvertor: convertor.NewImageconvertor(
				cfgs.FileService.FileConstraints),
		},
		querier,
		cfgs,
	)
	if err != nil {
		tele.Fatalf("failed to initialize media application: %v", err)
	}

	w := dbservice.NewWorker(querier)

	app.StartVariantWorker(ctx, cfgs.FileService.VariantWorkerInterval)
	w.StartStaleFilesWorker(ctx, cfgs.DB.StaleFilesWorkerInterval)

	service := &handler.MediaHandler{
		Application: app,
		// Configs:     cfgs.Server,
	}

	tele.Info(ctx, "Running gRpc service...")
	startServerFunc, endServerFunc, err := gorpc.CreateGRpcServer[media.MediaServiceServer](
		media.RegisterMediaServiceServer,
		service,
		cfgs.Server.GrpcServerPort,
		ct.CommonKeys(),
	)
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

	// wait here for process termination signal to initiate graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	tele.Info(ctx, "Shutting down server...")
	endServerFunc()
	tele.Info(ctx, "Server stopped")
	return nil
}

func getConfigs() configs.Config {
	return configs.Config{
		Server: configs.Server{
			GrpcServerPort: os.Getenv("GRPC_SERVER_PORT"),
		},
		DB: configs.Db{
			URL:                      os.Getenv("DATABASE_URL"),
			StaleFilesWorkerInterval: 1 * time.Hour,
		},
		FileService: configs.FileService{
			Buckets: configs.Buckets{
				Originals: "uploads-originals",
				Variants:  "uploads-variants",
			},
			VariantWorkerInterval: 30 * time.Second,
			FileConstraints: configs.FileConstraints{
				MaxImageUpload: 5 << 20, // 5MB
				MaxWidth:       4096,
				MaxHeight:      4096,
				AllowedMIMEs: map[string]bool{
					"image/jpeg": true,
					"image/jpg":  true,
					"image/png":  true,
					"image/gif":  true,
					"image/webp": true,
				},
				AllowedExt: map[string]bool{
					".jpg":  true,
					".jpeg": true,
					".png":  true,
					".gif":  true,
					".webp": true,
				},
			},
			Endpoint:       os.Getenv("MINIO_ENDPOINT"),
			PublicEndpoint: os.Getenv("MINIO_PUBLIC_ENDPOINT"),
			AccessKey:      os.Getenv("MINIO_ACCESS_KEY"),
			Secret:         os.Getenv("MINIO_SECRET_KEY"),
		},
		Tele: configs.Tele{
			EnableDebugLogs:           true,
			SimplePrint:               true,
			TelemetryCollectorAddress: os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
			OtelResourceAttributes:    os.Getenv("OTEL_RESOURCE_ATTRIBUTES"),
		},
	}
}
