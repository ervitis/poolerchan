package poolerchan

import (
	"context"
	"log"
)

type ConfigOption func(*Config)

func WithNumberOfJobs(nJobs int) ConfigOption {
	return func(config *Config) {
		config.numberOfJobs = nJobs
	}
}

func WithNumberOfWorkers(nWorkers int) ConfigOption {
	return func(config *Config) {
		config.numberOfWorkers = nWorkers
	}
}

func WithContext(ctx context.Context) ConfigOption {
	return func(config *Config) {
		config.context = ctx
	}
}

func WithLogger(logger *log.Logger) ConfigOption {
	return func(config *Config) {
		config.logger = logger
	}
}
