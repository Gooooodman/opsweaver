package logging

import (
	"io"
	"log/slog"
	"os"
	"strings"
)

type Options struct {
	Level   slog.Level
	Service string
	Writer  io.Writer
}

const maskedValue = "***"

var sensitiveKeys = map[string]struct{}{
	"token":                  {},
	"password":               {},
	"authorization":          {},
	"secret":                 {},
	"api_key":                {},
	"apikey":                 {},
	"dsn":                    {},
	"master_key":             {},
	"master_key_base64":      {},
	"internal_service_token": {},
}

func New(opts Options) *slog.Logger {
	writer := opts.Writer
	if writer == nil {
		writer = os.Stderr
	}

	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level:       opts.Level,
		ReplaceAttr: maskSensitiveAttr,
	})

	logger := slog.New(handler)
	if opts.Service != "" {
		logger = logger.With(slog.String("service", opts.Service))
	}
	return logger
}

func maskSensitiveAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) == 0 {
		switch attr.Key {
		case slog.TimeKey, slog.LevelKey, slog.MessageKey, "service":
			return attr
		}
	}
	if isSensitiveKey(attr.Key) {
		return slog.String(attr.Key, maskedValue)
	}
	return attr
}

func isSensitiveKey(key string) bool {
	_, ok := sensitiveKeys[strings.ToLower(key)]
	return ok
}
