package config

import "os"

type Config struct {
	MySQLDSN string
	HTTPAddr string
}

func Load() Config {
	return Config{
		MySQLDSN: env("MINIMALL_PRODUCT_MYSQL_DSN", "minimall:123456@tcp(127.0.0.1:3306)/minimall_product?parseTime=true"),
		HTTPAddr: env("MINIMALL_PRODUCT_HTTP_ADDR", ":8082"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
