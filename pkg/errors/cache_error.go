package errors

const (
	CodeCacheNotFound    = "err.cache.not_found"
	CodeCacheUnavailable = "err.cache.unavailable"
	CodeCacheSetFailed   = "err.cache.set_failed"
)

const (
	ErrFailedToSetGlobalCache = "Failed to set global cache"
	ErrFailedToSetCache       = "Failed to set cache"
	ErrFailedToGetCache       = "Failed to get cache"
	ErrCacheKeyNotFound       = "Cache key not found"
	ErrFailedToDeleteCache    = "Failed to delete cache"
	ErrFailedToClearCache     = "Failed to clear cache"
	ErrFailedToCloseCache     = "Failed to close cache connection"
)
