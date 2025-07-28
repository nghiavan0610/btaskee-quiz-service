package utils

func GetPointeValue[T any](t *T) T {
	var v T
	if t == nil {
		return v
	}

	return *t
}
