package impl

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	commonsCtx "gitlab.kenda.com.tw/kenda/commons/v2/utils/context"
)

// Logger definition.
type Logger struct {
	logLevel logger.LogLevel
}

// newLogger returns a new logger with log level Info.
func newLogger() logger.Interface {
	return &Logger{logLevel: logger.Info}
}

// LogMode resets log level.
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info logs formatted information message with corresponding message data.
func (l Logger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Info {
		commonsCtx.Logger(ctx).Info(fmt.Sprintf(msg, data...),
			zap.String("caller", utils.FileWithLineNum()))
	}
}

// Warn logs formatted warning message with corresponding message data.
func (l Logger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Warn {
		commonsCtx.Logger(ctx).Warn(fmt.Sprintf(msg, data...),
			zap.String("caller", utils.FileWithLineNum()))
	}
}

// Error logs formatted error message with corresponding message data.
func (l Logger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel >= logger.Error {
		commonsCtx.Logger(ctx).Error(fmt.Sprintf(msg, data...),
			zap.String("caller", utils.FileWithLineNum()))
	}
}

// Trace writes the tracing logs on console.
// Customize the logging info using the associated context logger.
func (l Logger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.String("caller", utils.FileWithLineNum()),
		zap.Duration("elapsed_time", elapsed),
		zap.String("sql", sql),
	}
	if rows != -1 {
		fields = append(fields, zap.Int64("rows_affected", rows))
	}
	if err != nil && l.logLevel >= logger.Error {
		fields = append(fields, zap.Error(err))
		commonsCtx.Logger(ctx).Error("tracing SQL..", fields...)
	} else {
		commonsCtx.Logger(ctx).Info("tracing SQL..", fields...)
	}
}
