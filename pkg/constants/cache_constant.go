package constants

type CachePrefix string
type CacheModule string
type CacheSuffix string

const (
	// Prefixes
	CachePrefixAccessToken  CachePrefix = "ACCESS_TOKEN"
	CachePrefixRefreshToken CachePrefix = "REFRESH_TOKEN"
	// ...add more as needed

	// Modules
	CacheModuleUser CacheModule = "USER"
	CacheModuleAuth CacheModule = "AUTH"
	CacheModuleRole CacheModule = "ROLE"
	// ...add more as needed

	// Suffixes
	// ...add more as needed
)
