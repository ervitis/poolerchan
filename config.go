package poolerchan

import (
	"log/slog"
	"os"
)

type Config struct {
	numberOfJobs    int
	numberOfWorkers int

	logger *slog.Logger
}

type ConfigOption func(*Config)

func defaultConfigPoolchan() *Config {
	return &Config{
		numberOfJobs:    defaultNumberOfJobs,
		numberOfWorkers: defaultNumberOfWorkers,
		logger:          slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
}

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

func WithLogger(logger *slog.Logger) ConfigOption {
	return func(config *Config) {
		config.logger = logger
	}
}
