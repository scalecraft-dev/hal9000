package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

func getEnv[T float64 | string | int | bool | time.Duration](key string, defaultVal T, required bool) T {
	val, ok := os.LookupEnv(key)
	if !ok {
		if !required {
			return defaultVal
		} else {
			log.Fatalf("missing required environment variable %s", key)
		}
	}

	var out T
	switch ptr := any(&out).(type) {
	case *string:
		{
			*ptr = val
		}
	case *int:
		{
			v, err := strconv.Atoi(val)
			if err != nil {
				return defaultVal
			}
			*ptr = v
		}
	case *bool:
		{
			v, err := strconv.ParseBool(val)
			if err != nil {
				return defaultVal
			}
			*ptr = v
		}
	case *time.Duration:
		{
			v, err := time.ParseDuration(val)
			if err != nil {
				return defaultVal
			}
			*ptr = v
		}
	case *float64:
		{
			v, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return defaultVal
			}
			*ptr = v
		}
	default:
		{
			log.Fatalf("unsupported type %T", out)
		}
	}

	return out
}
