package logger

import (
	"context"
	"time"

	"gorm.io/gorm/logger"
)

// gormLogger is a custom logger for GORM that redirects logs to our pkg/logger
type gormLogger struct {
	level logger.LogLevel
}

// NewGormLogger creates a new GORM logger that redirects to pkg/logger
func NewGormLogger() logger.Interface {
	return &gormLogger{
		level: logger.Warn, // Default to Warn level
	}
}

// LogMode sets the log level
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// Info logs messages at info level
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		Infof(msg, data...)
	}
}

// Warn logs messages at warn level
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		Warnf(msg, data...)
	}
}

// Error logs messages at error level
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		Errorf(msg, data...)
	}
}

// Trace logs SQL queries
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil && l.level >= logger.Error {
		if rows == -1 {
			Errorf("[GORM] [%.3fms] [rows:%v] %s, error: %v", float64(elapsed.Nanoseconds())/1e6, "-", sql, err)
		} else {
			Errorf("[GORM] [%.3fms] [rows:%v] %s, error: %v", float64(elapsed.Nanoseconds())/1e6, rows, sql, err)
		}
		return
	}

	if l.level >= logger.Warn {
		slowThreshold := 200 * time.Millisecond
		if elapsed > slowThreshold {
			if rows == -1 {
				Warnf("[GORM] [SLOW SQL >= %v] [%.3fms] [rows:%v] %s", slowThreshold, float64(elapsed.Nanoseconds())/1e6, "-", sql)
			} else {
				Warnf("[GORM] [SLOW SQL >= %v] [%.3fms] [rows:%v] %s", slowThreshold, float64(elapsed.Nanoseconds())/1e6, rows, sql)
			}
			return
		}
	}

	if l.level >= logger.Info {
		if rows == -1 {
			Infof("[GORM] [%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, "-", sql)
		} else {
			Infof("[GORM] [%.3fms] [rows:%v] %s", float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
