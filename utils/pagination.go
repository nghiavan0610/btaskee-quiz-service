package utils

func CalculateOffset(page, limit int32) (int32, int32) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return (page - 1) * limit, limit
}

func CalculateLimit(limit int32) int32 {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit
}
