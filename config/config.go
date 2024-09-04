package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	MongoDBURI         string
	TracingExporterURL string
}

func LoadEnvConfig(env string) (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	cfg.MongoDBURI = os.Getenv("MONGODB_URI")
	cfg.TracingExporterURL = os.Getenv("TRACING_EXPORTER_URL")
	if env != "local" {
		cfg.TracingExporterURL = strings.Replace(cfg.TracingExporterURL, "localhost", "jaeger", 1)
		cfg.MongoDBURI = strings.Replace(cfg.MongoDBURI, "localhost", "mongo", 1)
	}
	return cfg, nil
}
