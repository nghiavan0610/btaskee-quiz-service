package helpers

import (
	"crypto/rand"
	"math/big"
	"strings"
)

func GenerateJoinCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const codeLength = 6

	result := make([]byte, codeLength)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

func ValidateJoinCode(code string) bool {
	if len(code) != 6 {
		return false
	}

	code = strings.ToUpper(code)
	for _, char := range code {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return false
		}
	}

	return true
}

func NormalizeJoinCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
