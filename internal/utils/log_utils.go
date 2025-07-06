package utils

import (
	"context"
	"fmt"
	"github.com/DaminduDilsara/web-analyzer/configurations"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"strings"
)

type Field struct {
	Key   string
	Value interface{}
}

func SetLogFile(logPrefix string) Field {
	return Field{
		Key:   "file",
		Value: logPrefix,
	}
}

type LoggerInterface interface {
	Info(msg string, tags ...Field)
	InfoWithContext(ctx context.Context, msg string, tags ...Field)
	Error(msg string, err error, tags ...Field)
	ErrorWithContext(ctx context.Context, msg string, err error, tags ...Field)
	Fatal(msg string, err error, tags ...Field)
	FatalWithContext(ctx context.Context, msg string, err error, tags ...Field)
	Debug(msg string, tags ...Field)
	DebugWithContext(ctx context.Context, msg string, tags ...Field)
}

type logger struct {
	log       *zap.Logger
	logConfig *configurations.LogConfigurations
}

func InitLogger(appName string, logConfig *configurations.LogConfigurations) LoggerInterface {
	logFilePath := logConfig.LogFilePath
	if logFilePath == "" {
		logFilePath = "./logs"
	}

	logFile := fmt.Sprintf("%s/%s.log", logFilePath, appName)

	lumberjackLogger := &lumberjack.Logger{ // log file rotating configs
		Filename:   logFile,
		MaxSize:    100, // megabytes
		MaxBackups: 7,
		MaxAge:     1,    // days
		Compress:   true, // optional
	}

	fileWriter := zapcore.AddSync(lumberjackLogger)
	consoleWriter := zapcore.AddSync(os.Stdout)

	multiWriter := zapcore.NewMultiWriteSyncer(fileWriter, consoleWriter)

	loglevel := getLevel(logConfig.LogLevel)
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		TimeKey:     "timestamp",
		EncodeTime:  zapcore.ISO8601TimeEncoder,
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	}

	log := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		multiWriter,
		zap.NewAtomicLevelAt(loglevel),
	), zap.AddCaller())

	log = log.With(
		zap.String("ApplicationName", appName),
	)

	return &logger{
		log:       log,
		logConfig: logConfig,
	}
}

// InitConsoleLogger - this method is only for unit tests
func InitConsoleLogger() LoggerInterface {
	consoleWriter := zapcore.AddSync(os.Stdout)

	loglevel := zap.InfoLevel
	logEncoderConfig := zapcore.EncoderConfig{
		MessageKey:  "message",
		LevelKey:    "level",
		TimeKey:     "timestamp",
		EncodeTime:  zapcore.ISO8601TimeEncoder,
		EncodeLevel: zapcore.LowercaseLevelEncoder,
	}

	log := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(logEncoderConfig),
		consoleWriter,
		zap.NewAtomicLevelAt(loglevel),
	), zap.AddCaller())

	return &logger{
		log:       log,
		logConfig: nil,
	}
}

func (log logger) Info(msg string, tags ...Field) {
	log.log.Info(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) InfoWithContext(ctx context.Context, msg string, tags ...Field) {
	tags = append(tags, Field{"context", fmt.Sprintf("%v", ctx)})
	tags = append(tags, Field{"requestId", getRequestIdFromContext(ctx)})
	log.log.Info(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) Error(msg string, err error, tags ...Field) {
	msg = fmt.Sprintf("%s - ERROR - %v", msg, err)
	log.log.Error(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) ErrorWithContext(ctx context.Context, msg string, err error, tags ...Field) {
	tags = append(tags, Field{"context", fmt.Sprintf("%v", ctx)})
	tags = append(tags, Field{"requestId", getRequestIdFromContext(ctx)})
	msg = fmt.Sprintf("%s - ERROR - %v", msg, err)
	log.log.Error(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) Fatal(msg string, err error, tags ...Field) {
	msg = fmt.Sprintf("%s - FATAL - %v", msg, err)
	log.log.Fatal(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) FatalWithContext(ctx context.Context, msg string, err error, tags ...Field) {
	tags = append(tags, Field{"context", fmt.Sprintf("%v", ctx)})
	tags = append(tags, Field{"requestId", getRequestIdFromContext(ctx)})
	msg = fmt.Sprintf("%s - FATAL - %v", msg, err)
	log.log.Fatal(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) Debug(msg string, tags ...Field) {
	log.log.Debug(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) DebugWithContext(ctx context.Context, msg string, tags ...Field) {
	tags = append(tags, Field{"context", fmt.Sprintf("%v", ctx)})
	tags = append(tags, Field{"requestId", getRequestIdFromContext(ctx)})
	log.log.Debug(msg, log.fieldToZapField(tags...)...)
	log.log.Sync()
}

func (log logger) fieldToZapField(tags ...Field) []zap.Field {
	zapFields := make([]zap.Field, 0)
	for _, tag := range tags {
		zapFields = append(zapFields, zap.Any(tag.Key, tag.Value))
	}
	return zapFields
}

func getRequestIdFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("requestId").(string); ok {
		return requestID
	}
	return ""
}

func getLevel(logLevel string) zapcore.Level {
	switch strings.ToLower(strings.TrimSpace(logLevel)) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	}
	return zap.InfoLevel // if no log level given, set default as info level
}
