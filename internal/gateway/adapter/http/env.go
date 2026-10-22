package httpapi

import "os"

func getEnv(key string) string {
	return os.Getenv(key)
}
