package configs

import "time"

type Config struct {
	Server      Server
	DB          Db
	FileService FileService
	Clients     Clients
	Tele        Tele
}

type FileService struct {
	Endpoint              string `env:"MINIO_ENDPOINT"`
	PublicEndpoint        string `env:"MINIO_PUBLIC_ENDPOINT"` // used only for dev. Should be empty for production
	AccessKey             string `env:"MINIO_ACCESS_KEY"`
	Secret                string `env:"MINIO_SECRET_KEY"`
	Buckets               Buckets
	FileConstraints       FileConstraints
	VariantWorkerInterval time.Duration
}

// !!! Only use string types here !!!
type Buckets struct {
	Originals string
	Variants  string
}

type FileConstraints struct {
	MaxImageUpload int64
	AllowedMIMEs   map[string]bool
	AllowedExt     map[string]bool
	MaxWidth       int
	MaxHeight      int
}

type Server struct {
	GrpcServerPort string `env:"GRPC_SERVER_PORT"`
}

type Clients struct {
}

type Db struct {
	URL                      string `env:"DATABASE_URL"`
	StaleFilesWorkerInterval time.Duration
}

type Tele struct {
	EnableDebugLogs           bool   `env:"ENABLE_DEBUG_LOGS"`
	SimplePrint               bool   `env:"ENABLE_SIMPLE_PRINT"`
	OtelResourceAttributes    string `end:"OTEL_RESOURCE_ATTRIBUTES"`
	TelemetryCollectorAddress string `env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}
