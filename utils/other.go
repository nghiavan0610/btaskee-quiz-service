package utils

import (
	"encoding/json"
	"os"
	"strings"
)

func IsEmpty[T any](arr []T) bool {
	if arr == nil {
		return true
	}

	return len(arr) == 0
}

func IsBlank(v *string) bool {
	if v == nil {
		return true
	}

	return len(*v) == 0
}

func WriteFile(name string, data any) {
	b, _ := json.Marshal(data)
	_ = os.WriteFile(name, b, 0o644)
}

func Retry[T any](f func() (T, error), maxRetry uint8) (T, error) {
	var (
		result T
		err    error
	)
	for range maxRetry {
		result, err = f()
		if err == nil {
			return result, nil
		}
	}

	return result, err
}

func QueryStringInit(prefix string) *strings.Builder {
	var result strings.Builder
	if prefix != "" {
		result.WriteString(prefix)
	} else {
		result.WriteString("1=1 ")
	}

	return &result
}
