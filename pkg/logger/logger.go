package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog/pkgerrors"
)

// OutputStyle defines the format of log output
type OutputStyle string

const (
	// OutputStyleJSON outputs logs as JSON (default for services)
	OutputStyleJSON OutputStyle = "json"
	// OutputStyleConsole outputs logs in a human-readable format with colors (default for CLI)
	OutputStyleConsole OutputStyle = "console"
)

// Config holds the logger configuration
type Config struct {
	outputStyle     OutputStyle
	level           zerolog.Level
	timeFieldFormat string
	output          io.Writer
	addCaller       bool
	addTimestamp    bool
}

// Option is a functional option for configuring the logger
type Option func(*Config)

// WithOutputStyle sets the output style for the logger
func WithOutputStyle(style OutputStyle) Option {
	return func(c *Config) {
		c.outputStyle = style
	}
}

// WithLevel sets the log level
func WithLevel(level zerolog.Level) Option {
	return func(c *Config) {
		c.level = level
	}
}

// WithTimeFieldFormat sets the time field format
func WithTimeFieldFormat(format string) Option {
	return func(c *Config) {
		c.timeFieldFormat = format
	}
}

// WithOutput sets the output writer
func WithOutput(w io.Writer) Option {
	return func(c *Config) {
		c.output = w
	}
}

// WithCaller adds caller information to logs
func WithCaller(add bool) Option {
	return func(c *Config) {
		c.addCaller = add
	}
}

// WithTimestamp adds timestamp to logs
func WithTimestamp(add bool) Option {
	return func(c *Config) {
		c.addTimestamp = add
	}
}

// defaultConfig returns a config with default values
func defaultConfig() *Config {
	return &Config{
		outputStyle:     OutputStyleJSON,
		level:           zerolog.InfoLevel,
		timeFieldFormat: time.RFC3339,
		output:          os.Stdout,
		addCaller:       false,
		addTimestamp:    true,
	}
}

// InitialiseLogger sets up the logger with custom options
func InitialiseLogger(opts ...Option) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// Set global time format and error marshaler
	//nolint:reassign // intentional configuration of zerolog global settings
	zerolog.TimeFieldFormat = cfg.timeFieldFormat
	//nolint:reassign // intentional configuration of zerolog global settings
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Configure output writer based on style
	var writer io.Writer
	switch cfg.outputStyle {
	case OutputStyleConsole:
		writer = zerolog.ConsoleWriter{
			Out:        cfg.output,
			TimeFormat: time.RFC3339,
		}
	default: // OutputStyleJSON
		writer = cfg.output
	}

	// Create logger with configuration
	logger := zerolog.New(writer)

	if cfg.addTimestamp {
		logger = logger.With().Timestamp().Logger()
	}

	if cfg.addCaller {
		logger = logger.With().Caller().Logger()
	}

	// Set as global logger
	//nolint:reassign // intentional configuration of zerolog global logger
	log.Logger = logger

	// Set global log level
	zerolog.SetGlobalLevel(cfg.level)
}

// InitCLILogger initializes a logger pre-configured for CLI applications
// Uses console output with colors and compact format for better readability
func InitCLILogger(opts ...Option) {
	defaultOpts := []Option{
		WithOutputStyle(OutputStyleConsole),
		WithLevel(zerolog.InfoLevel),
		WithTimestamp(true),
		WithCaller(false),
	}

	// User options override defaults
	combinedOpts := make([]Option, 0, len(defaultOpts)+len(opts))
	combinedOpts = append(combinedOpts, defaultOpts...)
	combinedOpts = append(combinedOpts, opts...)
	InitialiseLogger(combinedOpts...)
}

// InitServiceLogger initializes a logger pre-configured for service applications
// Uses JSON output for structured logging, includes caller info for debugging
func InitServiceLogger(opts ...Option) {
	defaultOpts := []Option{
		WithOutputStyle(OutputStyleJSON),
		WithLevel(zerolog.InfoLevel),
		WithTimestamp(true),
		WithCaller(true),
	}

	// User options override defaults
	combinedOpts := make([]Option, 0, len(defaultOpts)+len(opts))
	combinedOpts = append(combinedOpts, defaultOpts...)
	combinedOpts = append(combinedOpts, opts...)
	InitialiseLogger(combinedOpts...)
}

// UpdateLoggerContext adds a key-value and returns a new context with the updated logger.
func UpdateLoggerContext(ctx context.Context, key, value string) context.Context {
	logger := log.Ctx(ctx).With().Str(key, value).Logger()
	return logger.WithContext(ctx)
}

// SetLogLevel sets the global log level from a string
func SetLogLevel(level string) {
	l, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to parse log level")
	}
	zerolog.SetGlobalLevel(l)
}
