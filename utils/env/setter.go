package env

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func setFieldFromEnv(field reflect.Value, envKey, defaultVal string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(getEnv(envKey, defaultVal))
	case reflect.Bool:
		defaultBool, _ := strconv.ParseBool(defaultVal)
		field.SetBool(getEnvBool(envKey, defaultBool))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		defaultInt, _ := strconv.ParseInt(defaultVal, 10, 64)
		field.SetInt(getEnvInt(envKey, defaultInt))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		defaultUint, _ := strconv.ParseUint(defaultVal, 10, 64)
		setUintField(field, envKey, defaultUint)
	case reflect.Float32, reflect.Float64:
		defaultFloat, _ := strconv.ParseFloat(defaultVal, 64)
		field.SetFloat(getEnvFloat(envKey, defaultFloat))
	case reflect.Slice:
		setSliceField(field, envKey, defaultVal)
	default:
		setDurationField(field, envKey, defaultVal)
	}
}

func setUintField(field reflect.Value, envKey string, defaultVal uint64) {
	if value := os.Getenv(envKey); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
			field.SetUint(parsed)
			return
		}
	}
	field.SetUint(defaultVal)
}

func setDurationField(field reflect.Value, envKey, defaultVal string) {
	if field.Type() == reflect.TypeFor[time.Duration]() {
		defaultDuration, _ := time.ParseDuration(defaultVal)
		field.Set(reflect.ValueOf(getEnvDuration(envKey, defaultDuration)))
	}
}

func setSliceField(field reflect.Value, envKey, defaultVal string) {
	if field.Type().Elem().Kind() == reflect.String {
		var defaultSlice []string
		if defaultVal != "" {
			parts := strings.SplitSeq(defaultVal, ",")
			for part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					defaultSlice = append(defaultSlice, trimmed)
				}
			}
		}
		result := getEnvStringSlice(envKey, defaultSlice)
		field.Set(reflect.ValueOf(result))
	}
}

func setFieldDefault(field reflect.Value, defaultVal string) {
	switch field.Kind() {
	case reflect.String:
		field.SetString(defaultVal)
	case reflect.Bool:
		if defaultBool, err := strconv.ParseBool(defaultVal); err == nil {
			field.SetBool(defaultBool)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if defaultInt, err := strconv.ParseInt(defaultVal, 10, 64); err == nil {
			field.SetInt(defaultInt)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if defaultUint, err := strconv.ParseUint(defaultVal, 10, 64); err == nil {
			field.SetUint(defaultUint)
		}
	case reflect.Float32, reflect.Float64:
		if defaultFloat, err := strconv.ParseFloat(defaultVal, 64); err == nil {
			field.SetFloat(defaultFloat)
		}
	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String && defaultVal != "" {
			parts := strings.Split(defaultVal, ",")
			result := make([]string, 0, len(parts))
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
			field.Set(reflect.ValueOf(result))
		}
	default:
		if field.Type() == reflect.TypeFor[time.Duration]() {
			if defaultDuration, err := time.ParseDuration(defaultVal); err == nil {
				field.Set(reflect.ValueOf(defaultDuration))
			}
		}
	}
}
