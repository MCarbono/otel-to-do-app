package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type config struct {
	MongoDBURI         string
	TracingExporterURL string
	ApiServerPort      string
	TodoServerPort     string
}

func LoadEnvConfig(env string) (config, error) {
	var cfg config
	err := godotenv.Load()
	if err != nil {
		return cfg, err
	}
	cfg.ApiServerPort = os.Getenv("API_SERVER_PORT")
	cfg.TodoServerPort = os.Getenv("TODO_SERVER_PORT")
	cfg.MongoDBURI = os.Getenv("MONGODB_URI")
	cfg.TracingExporterURL = os.Getenv("TRACING_EXPORTER_URL")
	if env != "local" {
		cfg.TracingExporterURL = strings.Replace(cfg.TracingExporterURL, "localhost", "jaeger", 1)
		cfg.MongoDBURI = strings.Replace(cfg.MongoDBURI, "localhost", "mongo", 1)
	}
	return cfg, nil
}
