package errors

const (
	CodeInternal           = "err.internal"
	CodeBadRequest         = "err.bad_request"
	CodeValidation         = "err.validation"
	CodeNotFound           = "err.not_found"
	CodeUnauthorized       = "err.unauthorized"
	CodeForbidden          = "err.forbidden"
	CodeConflict           = "err.conflict"
	CodeAlreadyExists      = "err.already_exists"
	CodeRateLimitExceeded  = "err.rate_limit_exceeded"
	CodeServiceUnavailable = "err.service_unavailable"
	CodeDBError            = "err.db_error"
)

const (
	ErrRouteNotFound         = "The requested route was not found"
	ErrInvalidInput          = "Invalid input provided"
	ErrInternal              = "An internal error occurred"
	ErrConflict              = "Conflict occurred"
	ErrServiceUnavailable    = "Service is currently unavailable"
	ErrRateLimitExceeded     = "Rate limit exceeded"
	ErrHTTP                  = "HTTP error occurred"
	ErrUnexpected            = "Unexpected error occurred"
	ErrTooManyRequests       = "Too many requests, please try again later"
	ErrInvalidTimeFormat     = "Invalid time format provided"
	ErrInvalidPathParams     = "Invalid path parameters"
	ErrInvalidQueryParams    = "Invalid query parameters"
	ErrInvalidBodyParams     = "Invalid body parameters"
	ErrInvalidPayloadParams  = "Invalid payload parameters"
	ErrFailedToUnmarshalJSON = "Failed to unmarshal JSON"
	ErrFailedToMarshalJSON   = "Failed to marshal JSON"
)
