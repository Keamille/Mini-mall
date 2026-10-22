package config

import "os"

type Config struct {
	MySQLDSN string
	HTTPAddr string
	GRPCAddr string
}

func Load() Config {
	return Config{
		MySQLDSN: env("MINIMALL_COUPON_MYSQL_DSN", "minimall:123456@tcp(127.0.0.1:3306)/minimall_coupon?parseTime=true"),
		HTTPAddr: env("MINIMALL_COUPON_HTTP_ADDR", ":8086"),
		GRPCAddr: env("MINIMALL_COUPON_GRPC_ADDR", ":9006"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
