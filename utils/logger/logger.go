package logger

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const prefixWidth = 15

func getTimestamp() string {
	return time.Now().Format(time.RFC3339)
}

func getLevelColor(level LogLevel) string {
	switch level {
	case Info:
		return LevelColorInfo
	case Warn:
		return LevelColorWarn
	case Error:
		return LevelColorError
	case Debug:
		return LevelColorDebug
	case Success:
		return LevelColorSuccess
	default:
		return LevelColorInfo
	}
}

func getMessageColor(level LogLevel) string {
	switch level {
	case Info:
		return MessageColorInfo
	case Warn:
		return MessageColorWarn
	case Error:
		return MessageColorError
	case Debug:
		return MessageColorDebug
	case Success:
		return MessageColorSuccess
	default:
		return MessageColorInfo
	}
}

func Log(message interface{}, options LogOptions) {
	var builder strings.Builder

	if options.Timestamp {
		builder.WriteString(Gray)
		builder.WriteString(getTimestamp())
		builder.WriteString(Reset)
		builder.WriteString(" ")
	}

	builder.WriteString(getLevelColor(options.Level))
	builder.WriteString(" ")

	if options.Prefix != "" {
		totalWidth := len(options.Prefix)
		padding := ""

		if totalWidth < prefixWidth {
			padding = strings.Repeat(" ", prefixWidth-totalWidth)
		}

		builder.WriteString(Cyan)
		builder.WriteString("[")
		builder.WriteString(options.Prefix)
		builder.WriteString("]")
		builder.WriteString(Reset)
		builder.WriteString(padding)
		builder.WriteString(" ")
	}

	builder.WriteString(getMessageColor(options.Level))

	switch msg := message.(type) {
	case error:
		builder.WriteString(msg.Error())
	case string:
		builder.WriteString(msg)
	default:
		builder.WriteString(fmt.Sprintf("%v", msg))
	}

	builder.WriteString(Reset)
	builder.WriteString("\n")

	if options.Level == Error || options.Level == Warn {
		os.Stderr.WriteString(builder.String())
	} else {
		os.Stdout.WriteString(builder.String())
	}

	if options.Fatal {
		os.Exit(1)
	}
}
