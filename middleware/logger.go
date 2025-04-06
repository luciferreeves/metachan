package middleware

import (
	"fmt"
	"metachan/types"
	"metachan/utils/logger"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func HTTPLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		ip := c.IP()

		level := getLogLevel(status)
		messageColor := getMessageColor(level)

		// Make sure method is padded within the message itself
		paddedMethod := method
		if len(method) < 7 { // "DELETE" is 6 chars, add a margin
			paddedMethod = method + strings.Repeat(" ", 7-len(method))
		}

		// Format with consistent spacing and alignment
		message := fmt.Sprintf("%s %s%-3d%s %-15s %-10s %-s",
			paddedMethod,
			messageColor, status, types.Reset,
			"IP: "+ip,
			"TTR: "+formatDuration(duration),
			"Path: "+path,
		)

		logger.Log(message, types.LogOptions{
			Prefix: "HTTP",
			Level:  level,
		})

		return err
	}
}

func getLogLevel(status int) types.LogLevel {
	switch {
	case status >= 500:
		return types.Error
	case status >= 400:
		return types.Warn
	case status >= 300:
		return types.Info
	case status >= 200:
		return types.Success
	default:
		return types.Info
	}
}

func getMessageColor(level types.LogLevel) string {
	switch level {
	case types.Info:
		return types.MessageColorInfo
	case types.Warn:
		return types.MessageColorWarn
	case types.Error:
		return types.MessageColorError
	case types.Debug:
		return types.MessageColorDebug
	case types.Success:
		return types.MessageColorSuccess
	default:
		return types.MessageColorInfo
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return strconv.FormatInt(d.Nanoseconds(), 10) + "ns"
	} else if d < time.Millisecond {
		return strconv.FormatInt(d.Nanoseconds()/1000, 10) + "µs"
	} else if d < time.Second {
		return strconv.FormatFloat(float64(d.Nanoseconds())/float64(time.Millisecond), 'f', 3, 64) + "ms"
	} else {
		return strconv.FormatFloat(float64(d.Nanoseconds())/float64(time.Second), 'f', 3, 64) + "s"
	}
}
