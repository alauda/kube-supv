package log

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func init() {
	Logger()
}

var logger *zap.Logger

// Logger return the logger
func Logger() *zap.Logger {
	if logger == nil {
		var err error
		logger, err = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}.Build(
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zapcore.DPanicLevel),
		)
		if err != nil {
			panic(err)
		}
	}
	return logger
}

func Fatalf(format string, v ...interface{}) {
	defer logger.Sync()
	logger.Fatal(fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	defer logger.Sync()
	logger.Error(fmt.Sprintf(format, v...))
}

func Infof(format string, v ...interface{}) {
	defer logger.Sync()
	logger.Info(fmt.Sprintf(format, v...))
}

func Debugf(format string, v ...interface{}) {
	defer logger.Sync()
	logger.Debug(fmt.Sprintf(format, v...))
}

func Warnf(format string, v ...interface{}) {
	defer logger.Sync()
	logger.Warn(fmt.Sprintf(format, v...))
}

func DPanicf(format string, v ...interface{}) {
	defer func() {
		logger.Sync()
		recover()
	}()
	logger.DPanic(fmt.Sprintf(format, v...))
}
