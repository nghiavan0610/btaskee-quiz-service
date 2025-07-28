package constants

type KeyType string

const (
	KEY_AUTH_USER      KeyType = "x-authorization-key"
	KEY_JWT_CLAIMS     KeyType = "x-jwt-claims"
	KEY_CURRENT_TRAN   KeyType = "current-transaction-key"
	KEY_CORRELATION_ID KeyType = "correlation-id"
)

type ReqSourceType string

const (
	KEY_REQ_PATH_PARAMS    ReqSourceType = "req.path.params"
	KEY_REQ_QUERY_PARAMS   ReqSourceType = "req.query.params"
	KEY_REQ_BODY_PARAMS    ReqSourceType = "req.body.params"
	KEY_REQ_PAYLOAD_PARAMS ReqSourceType = "req.payload.params"
)

const (
	// ParallelProcessingThreshold is the minimum number of items to trigger parallel processing
	ParallelProcessingThreshold = 50

	// SmallDatasetThreshold is the threshold below which sequential processing is used
	SmallDatasetThreshold = 10
)

// Database query constants
const (
	// DatabaseQueryTimeout is the timeout for database queries
	DatabaseQueryTimeout = 30 // seconds
)
