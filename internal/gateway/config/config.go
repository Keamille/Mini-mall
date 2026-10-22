package config

import "os"

func HTTPAddr() string {
	if v := os.Getenv("MINIMALL_GATEWAY_HTTP_ADDR"); v != "" {
		return v
	}
	return ":8080"
}
