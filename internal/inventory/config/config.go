package config

import "os"

type Config struct {
	MySQLDSN string
	HTTPAddr string
	GRPCAddr string
}

func Load() Config {
	return Config{
		MySQLDSN: env("MINIMALL_INVENTORY_MYSQL_DSN", "minimall:123456@tcp(127.0.0.1:3306)/minimall_inventory?parseTime=true"),
		HTTPAddr: env("MINIMALL_INVENTORY_HTTP_ADDR", ":8088"),
		GRPCAddr: env("MINIMALL_INVENTORY_GRPC_ADDR", ":9008"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
