package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const prefixWidth = 15

var (
	loggerInstance *zap.Logger
	level          zap.AtomicLevel
)

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(Gray + t.Format(time.RFC3339) + Reset)
}

func levelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	switch l {
	case zapcore.DebugLevel:
		enc.AppendString(LevelColorDebug)
	case zapcore.WarnLevel:
		enc.AppendString(LevelColorWarn)
	case zapcore.ErrorLevel:
		enc.AppendString(LevelColorError)
	default:
		enc.AppendString(LevelColorInfo)
	}
}

func formatPrefix(prefix string) string {
	if prefix == "" {
		return ""
	}

	padding := ""
	if len(prefix) < prefixWidth {
		padding = strings.Repeat(" ", prefixWidth-len(prefix))
	}

	return Cyan + "[" + prefix + "]" + Reset + padding
}

func colorMessage(level LogLevel, msg string) string {
	switch level {
	case Debug:
		return MessageColorDebug + msg + Reset
	case Warn:
		return MessageColorWarn + msg + Reset
	case Error:
		return MessageColorError + msg + Reset
	case Success:
		return MessageColorSuccess + msg + Reset
	default:
		return MessageColorInfo + msg + Reset
	}
}

func Init() {
	level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:     "ts",
		LevelKey:    "level",
		MessageKey:  "msg",
		LineEnding:  "\n",
		EncodeTime:  timeEncoder,
		EncodeLevel: levelEncoder,
	}

	encoder := zapcore.NewConsoleEncoder(encoderCfg)

	stdout := zapcore.AddSync(os.Stdout)
	stderr := zapcore.AddSync(os.Stderr)

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, stdout, zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l < zapcore.WarnLevel && level.Enabled(l)
		})),
		zapcore.NewCore(encoder, stderr, zap.LevelEnablerFunc(func(l zapcore.Level) bool {
			return l >= zapcore.WarnLevel && level.Enabled(l)
		})),
	)

	loggerInstance = zap.New(core, zap.AddCaller())
}

func SetDebug(enabled bool) {
	if enabled {
		level.SetLevel(zapcore.DebugLevel)
	} else {
		level.SetLevel(zapcore.InfoLevel)
	}
}

func Debugf(prefix, format string, args ...any) {
	log(Debug, zapcore.DebugLevel, prefix, fmt.Sprintf(format, args...))
}

func Infof(prefix, format string, args ...any) {
	log(Info, zapcore.InfoLevel, prefix, fmt.Sprintf(format, args...))
}

func Successf(prefix, format string, args ...any) {
	log(Success, zapcore.InfoLevel, prefix, fmt.Sprintf(format, args...))
}

func Warnf(prefix, format string, args ...any) {
	log(Warn, zapcore.WarnLevel, prefix, fmt.Sprintf(format, args...))
}

func Errorf(prefix, format string, args ...any) {
	log(Error, zapcore.ErrorLevel, prefix, fmt.Sprintf(format, args...))
}

func Fatalf(prefix, format string, args ...any) {
	log(Error, zapcore.ErrorLevel, prefix, fmt.Sprintf(format, args...))
	os.Exit(1)
}

func log(levelLabel LogLevel, zapLevel zapcore.Level, prefix string, msg any) {
	if loggerInstance == nil {
		panic("logger.Init() was not called")
	}

	message := fmt.Sprint(msg)
	colored := colorMessage(levelLabel, message)
	fullMessage := formatPrefix(prefix) + colored

	loggerInstance.Log(zapLevel, fullMessage)
}
