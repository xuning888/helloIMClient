package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var globalLogger *zap.Logger

func InitLogger() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	absPath := filepath.Join(homeDir, ".helloIm", "log", "app.log")
	dir := filepath.Dir(absPath)
	if err2 := os.MkdirAll(dir, 0755); err2 != nil {
		return err2
	}
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   absPath,
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
	})

	// 创建编码器
	encoderConfig := DefaultZapLoggerConfig.EncoderConfig
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, fileWriter, DefaultZapLoggerConfig.Level),
		zapcore.NewCore(encoder, zapcore.AddSync(zapcore.Lock(os.Stderr)), DefaultZapLoggerConfig.Level),
	)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	zap.ReplaceGlobals(logger)
	globalLogger = logger
	return nil
}

var DefaultLevel = zap.InfoLevel

var DefaultZapLoggerConfig = zap.Config{
	Level:       zap.NewAtomicLevelAt(DefaultLevel),
	Development: true,
	Sampling: &zap.SamplingConfig{
		Initial:    100,
		Thereafter: 100,
	},

	Encoding: "console",

	// copied from "zap.NewProductionEncoderConfig" with some updates
	EncoderConfig: zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,

		// Custom EncodeTime function to ensure we match format and precision of historic capnslog timestamps
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},

		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	},

	OutputPaths:      []string{"stderr"},
	ErrorOutputPaths: []string{"stderr"},
}
