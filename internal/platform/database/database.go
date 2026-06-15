// Package database wires GORM-backed PostgreSQL connections for OpsWeaver
// services. Each caller is expected to inject a DSN; the package itself does
// not load configuration, so it can be reused by any service that owns a
// distinct database (for example opsweaver-server and opsweaver-gateway).
package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ErrNilDB is returned by Close when called with a nil *gorm.DB.
var ErrNilDB = errors.New("database: nil *gorm.DB")

const (
	defaultMaxOpenConns    = 10
	defaultMaxIdleConns    = 5
	defaultConnMaxLifetime = 30 * time.Minute
	defaultSlowThreshold   = 200 * time.Millisecond
)

// Options configures a database connection. DSN is required; all other fields
// fall back to safe defaults.
type Options struct {
	DSN             string
	Logger          *slog.Logger
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	SlowThreshold   time.Duration
}

// Open creates a *gorm.DB backed by the given PostgreSQL DSN, configures the
// connection pool, and verifies connectivity with a context-bound ping. The
// returned *gorm.DB carries a slog-backed logger if opts.Logger is non-nil; a
// silent logger is used otherwise.
func Open(ctx context.Context, opts Options) (*gorm.DB, error) {
	if opts.DSN == "" {
		return nil, errors.New("open: DSN is required")
	}

	cfg := &gorm.Config{
		Logger: buildGormLogger(opts.Logger, slowThresholdOrDefault(opts.SlowThreshold)),
	}

	db, err := gorm.Open(postgres.Open(opts.DSN), cfg)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("open: acquire underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(positiveOrDefault(opts.MaxOpenConns, defaultMaxOpenConns))
	sqlDB.SetMaxIdleConns(positiveOrDefault(opts.MaxIdleConns, defaultMaxIdleConns))
	sqlDB.SetConnMaxLifetime(durationOrDefault(opts.ConnMaxLifetime, defaultConnMaxLifetime))

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("open: ping database: %w", err)
	}

	return db, nil
}

// Close releases the underlying *sql.DB held by the *gorm.DB.
func Close(db *gorm.DB) error {
	if db == nil {
		return ErrNilDB
	}
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("close: acquire underlying sql.DB: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close: %w", err)
	}
	return nil
}

func positiveOrDefault(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func durationOrDefault(value, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return value
}

func slowThresholdOrDefault(value time.Duration) time.Duration {
	if value <= 0 {
		return defaultSlowThreshold
	}
	return value
}

func buildGormLogger(logger *slog.Logger, slowThreshold time.Duration) gormlogger.Interface {
	if logger == nil {
		return gormlogger.Discard
	}
	return &gormSlogLogger{
		logger:        logger,
		level:         gormlogger.Warn,
		slowThreshold: slowThreshold,
	}
}

// gormSlogLogger adapts a *slog.Logger to gorm's logger.Interface. The slog
// logger is expected to be created by the project's logging package so its
// ReplaceAttr-based sensitive-key masking applies uniformly across all log
// records. SQL text itself is not masked because it is not classified as a
// sensitive key by the logging package.
type gormSlogLogger struct {
	logger        *slog.Logger
	level         gormlogger.LogLevel
	slowThreshold time.Duration
}

// LogMode returns a copy of the logger with the requested verbosity.
func (l *gormSlogLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *gormSlogLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Info {
		return
	}
	l.logger.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf(msg, data...))
}

func (l *gormSlogLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Warn {
		return
	}
	l.logger.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf(msg, data...))
}

func (l *gormSlogLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level < gormlogger.Error {
		return
	}
	l.logger.LogAttrs(ctx, slog.LevelError, fmt.Sprintf(msg, data...))
}

func (l *gormSlogLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	attrs := []slog.Attr{
		slog.String("sql", sql),
		slog.Int64("rows", rows),
		slog.Duration("elapsed", elapsed),
	}

	switch {
	case err != nil && l.level >= gormlogger.Error:
		attrs = append(attrs, slog.String("error", err.Error()))
		l.logger.LogAttrs(ctx, slog.LevelError, "gorm query failed", attrs...)
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.level >= gormlogger.Warn:
		attrs = append(attrs, slog.Duration("slow_threshold", l.slowThreshold))
		l.logger.LogAttrs(ctx, slog.LevelWarn, "gorm slow query", attrs...)
	case l.level >= gormlogger.Info:
		l.logger.LogAttrs(ctx, slog.LevelInfo, "gorm query", attrs...)
	}
}
