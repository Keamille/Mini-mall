package config

import "os"

type Config struct {
	MySQLDSN string
	HTTPAddr string
	GRPCAddr string
	NATSURL  string
}

func Load() Config {
	return Config{
		MySQLDSN: env("MINIMALL_NOTIFICATION_MYSQL_DSN", "minimall:123456@tcp(127.0.0.1:3306)/minimall_notification?parseTime=true"),
		HTTPAddr: env("MINIMALL_NOTIFICATION_HTTP_ADDR", ":8087"),
		GRPCAddr: env("MINIMALL_NOTIFICATION_GRPC_ADDR", ":9009"),
		NATSURL:  env("MINIMALL_NOTIFICATION_NATS_URL", "nats://127.0.0.1:4222"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
