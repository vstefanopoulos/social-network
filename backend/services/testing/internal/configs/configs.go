package configs

import (
	configutil "social-network/shared/go/configs"
	tele "social-network/shared/go/telemetry"
)

type Configs struct {
	RedisAddr       string   `env:"REDIS_ADDR"`
	SentinelAddrs   []string `env:"SENTINEL_ADDRS"`
	RedisPassword   string   `env:"REDIS_PASSWORD"`
	RedisMasterName string   `env:"REDIS_MASTER"`
	RedisDB         int      `env:"REDIS_DB"`

	UsersGRPCAddr string `env:"USERS_GRPC_ADDR"`
	PostsGRPCAddr string `env:"POSTS_GRPC_ADDR"`
	ChatGRPCAddr  string `env:"CHAT_GRPC_ADDR"`
	MediaGRPCAddr string `env:"MEDIA_GRPC_ADDR"`

	HTTPAddr        string `env:"HTTP_ADDR"`
	ShutdownTimeout int    `env:"SHUTDOWN_TIMEOUT_SECONDS"`
}

func GetConfigs() Configs { // sensible defaults
	cfgs := Configs{
		RedisAddr:       "redis:6379",
		SentinelAddrs:   []string{"redis:26379"},
		RedisPassword:   "admin",
		RedisMasterName: "mymaster",
		RedisDB:         0,
		UsersGRPCAddr:   "users:50051",
		PostsGRPCAddr:   "posts:50051",
		ChatGRPCAddr:    "chat:50051",
		MediaGRPCAddr:   "media:50051",
		HTTPAddr:        "0.0.0.0:8081",
		ShutdownTimeout: 5,
	}

	// load environment variables if present
	_, err := configutil.LoadConfigs(&cfgs)
	if err != nil {
		tele.Fatalf("failed to load env variables into config struct: %v", err)
	}

	return cfgs
}
