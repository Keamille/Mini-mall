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
		MySQLDSN: env("MINIMALL_ORDER_MYSQL_DSN", "minimall:123456@tcp(127.0.0.1:3306)/minimall_order?parseTime=true"),
		HTTPAddr: env("MINIMALL_ORDER_HTTP_ADDR", ":8089"),
		GRPCAddr: env("MINIMALL_ORDER_GRPC_ADDR", ":9004"),
		NATSURL:  env("MINIMALL_ORDER_NATS_URL", "nats://127.0.0.1:4222"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
