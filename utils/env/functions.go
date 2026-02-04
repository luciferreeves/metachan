package env

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func getEnv(key, defaultVal string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int64) int64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseInt(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvFloat(key string, defaultVal float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultVal
}

func getEnvStringSlice(key string, defaultVal []string) []string {
	if value := os.Getenv(key); value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultVal
}

func Defaults[T any](config *T) *T {
	v := reflect.ValueOf(config)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return config
	}

	elem := v.Elem()
	t := elem.Type()
	newStruct := reflect.New(t)
	newElem := newStruct.Elem()

	for i := range elem.NumField() {
		field := newElem.Field(i)
		fieldType := t.Field(i)

		if !field.CanSet() {
			continue
		}

		defaultVal := fieldType.Tag.Get("default")
		if defaultVal == "" {
			continue
		}

		setFieldDefault(field, defaultVal)
	}

	return newStruct.Interface().(*T)
}
