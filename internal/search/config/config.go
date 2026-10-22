package config

import "os"

func OpenSearchURL() string {
	return env("MINIMALL_SEARCH_OPENSEARCH_URL", "http://127.0.0.1:9200")
}

func IndexAlias() string {
	return env("MINIMALL_SEARCH_INDEX", "minimall_products")
}

func HTTPAddr() string {
	return env("MINIMALL_SEARCH_HTTP_ADDR", ":8084")
}

func GRPCAddr() string {
	return env("MINIMALL_SEARCH_GRPC_ADDR", ":9007")
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
