package config

import "engflex-api/internal/utils"

// MinioConfig is the object-storage endpoint admin uploads write to. The
// same instance already serves reads via the web app's /api/media proxy
// (see apps/web/src/routes/api/media/$.ts) — this is the write side.
type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

func LoadMinioConfig() MinioConfig {
	return MinioConfig{
		Endpoint:  utils.GetEnv("MINIO_ENDPOINT", "http://localhost:9000"),
		AccessKey: utils.GetEnv("MINIO_ROOT_USER", ""),
		SecretKey: utils.GetEnv("MINIO_ROOT_PASSWORD", ""),
		UseSSL:    false,
	}
}
