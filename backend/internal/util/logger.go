package util

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"organiq/config"
)

// LoggerOptions controla a configuração adicional do logger.
type LoggerOptions struct {
	Service     string
	ExtraFields map[string]string
	IncludeHook bool
}

// ContextFields representa metadados opcionais anexados ao logger por request.
type ContextFields struct {
	RequestID string
	UserID    string
	Email     string
	IP        string
	TraceID   string
}

// InitLogger inicializa o zerolog global de forma consistente para todos os serviços.
// Retorna uma função de cleanup opcional (para fechar arquivos) e erro.
func InitLogger(cfg *config.Config, opts LoggerOptions) (func() error, error) {
	level, err := zerolog.ParseLevel(cfg.Logger.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	writer, closer, err := buildWriter(cfg)
	if err != nil {
		return nil, err
	}

	var hook zerolog.Hook
	if opts.IncludeHook {
		hook = staticFieldHook{fields: buildHookFields(cfg, opts)}
	}

	logger := zerolog.New(writer)
	if hook != nil {
		logger = logger.Hook(hook)
	}

	logger = logger.With().
		Timestamp().
		Str("service", opts.Service).
		Str("environment", cfg.Environment).
		Logger()

	log.Logger = logger

	return closer, nil
}

func buildWriter(cfg *config.Config) (io.Writer, func() error, error) {
	var outputs []io.Writer
	var cleanup func() error

	switch cfg.Logger.Output {
	case "file":
		if cfg.Logger.FilePath == "" {
			return nil, nil, fmt.Errorf("logger file output requires file path")
		}

		file, err := os.OpenFile(cfg.Logger.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to open log file: %w", err)
		}

		outputs = append(outputs, file)
		cleanup = func() error { return file.Close() }
	default:
		outputs = append(outputs, os.Stdout)
	}

	var writer io.Writer
	if len(outputs) == 1 {
		writer = outputs[0]
	} else {
		writer = io.MultiWriter(outputs...)
	}

	if cfg.Logger.Format == "console" {
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}
	} else {
		zerolog.TimeFieldFormat = time.RFC3339Nano
	}

	return writer, cleanup, nil
}

func buildHookFields(cfg *config.Config, opts LoggerOptions) map[string]string {
	fields := map[string]string{
		"environment": cfg.Environment,
	}
	if opts.Service != "" {
		fields["service"] = opts.Service
	}
	for k, v := range opts.ExtraFields {
		if k == "" || v == "" {
			continue
		}
		fields[k] = v
	}
	return fields
}

type staticFieldHook struct {
	fields map[string]string
}

func (h staticFieldHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	for k, v := range h.fields {
		e.Str(k, v)
	}
}

// WithContextFields retorna um novo context contendo logger com metadados adicionais.
func WithContextFields(ctx context.Context, fields ContextFields) context.Context {
	logger := log.Ctx(ctx)
	if logger == nil {
		l := log.Logger
		logger = &l
	}

	child := logger.With()
	if fields.RequestID != "" {
		child = child.Str("request_id", fields.RequestID)
	}
	if fields.UserID != "" {
		child = child.Str("user_id", fields.UserID)
	}
	if fields.Email != "" {
		child = child.Str("email", fields.Email)
	}
	if fields.IP != "" {
		child = child.Str("ip", fields.IP)
	}
	if fields.TraceID != "" {
		child = child.Str("trace_id", fields.TraceID)
	}

	loggerWithCtx := child.Logger()
	return loggerWithCtx.WithContext(ctx)
}

// LoggerFromContext garante um logger válido mesmo se contexto não possuir um.
func LoggerFromContext(ctx context.Context) *zerolog.Logger {
	if logger := log.Ctx(ctx); logger != nil {
		return logger
	}
	l := log.Logger
	return &l
}
