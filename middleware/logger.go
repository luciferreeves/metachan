package middleware

import (
	"fmt"
	"metachan/utils/logger"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func httpLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		ip := c.IP()

		// Pad method for alignment
		paddedMethod := method
		if len(method) < 7 {
			paddedMethod = method + strings.Repeat(" ", 7-len(method))
		}

		message := fmt.Sprintf(
			"%s %-3d %-15s %-10s %s",
			paddedMethod,
			status,
			"IP: "+ip,
			"TTR: "+formatDuration(duration),
			"Path: "+path,
		)

		logByStatus(status, "HTTP", message)

		return err
	}
}

func logByStatus(status int, prefix, message string) {
	switch {
	case status >= 500:
		logger.Errorf(prefix, "%s", message)
	case status >= 400:
		logger.Warnf(prefix, "%s", message)
	case status >= 300:
		logger.Infof(prefix, "%s", message)
	case status >= 200:
		logger.Successf(prefix, "%s", message)
	default:
		logger.Infof(prefix, "%s", message)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return strconv.FormatInt(d.Nanoseconds(), 10) + "ns"
	}
	if d < time.Millisecond {
		return strconv.FormatInt(d.Nanoseconds()/1_000, 10) + "µs"
	}
	if d < time.Second {
		return strconv.FormatFloat(
			float64(d.Nanoseconds())/float64(time.Millisecond),
			'f',
			3,
			64,
		) + "ms"
	}
	return strconv.FormatFloat(
		float64(d.Nanoseconds())/float64(time.Second),
		'f',
		3,
		64,
	) + "s"
}
