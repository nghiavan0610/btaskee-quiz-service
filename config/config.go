package config

import (
	"os"
	"strconv"
	"sync"
)

type Config struct {
	Server    ServerConfig
	JWT       JWTConfig
	RateLimit RateLimitConfig
	Redis     RedisConfig
	CORS      CORSConfig
	Database  DatabaseConfig
}

type ServerConfig struct {
	GoEnv          string
	ServiceName    string
	ServiceVersion string
	Host           string
	Port           string
}

type JWTConfig struct {
	AccessTokenSecret      string
	AccessTokenExpiration  int
	RefreshTokenSecret     string
	RefreshTokenExpiration int
}

type RateLimitConfig struct {
	RPS   float64
	Burst int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	CacheDB  int
	QueueDB  int
	Prefix   string
}

type CORSConfig struct {
	AllowOrigins     string
	AllowMethods     string
	AllowHeaders     string
	AllowCredentials bool
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	Schema          string
	SSLMode         string
	MaxConnections  int
	MinConnections  int
	MaxConnLifetime int // minutes
	MaxConnIdleTime int // minutes
}

var (
	config     *Config
	configOnce sync.Once
)

func ProvideConfig() *Config {
	configOnce.Do(func() {
		rateRPS, _ := strconv.ParseFloat(os.Getenv("RATE_LIMIT_RPS"), 64)
		rateBurst, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_BURST"))
		accessTokenExp, _ := strconv.Atoi(os.Getenv("JWT_ACCESS_TOKEN_EXPIRATION"))
		refreshTokenExp, _ := strconv.Atoi(os.Getenv("JWT_REFRESH_TOKEN_EXPIRATION"))
		redisCacheDB, _ := strconv.Atoi(os.Getenv("REDIS_CACHE_DB"))
		redisQueueDB, _ := strconv.Atoi(os.Getenv("REDIS_QUEUE_DB"))
		corsAllowCredentials, _ := strconv.ParseBool(os.Getenv("CORS_ALLOW_CREDENTIALS"))

		// Database config
		dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
		dbMaxConns, _ := strconv.Atoi(os.Getenv("DB_MAX_CONNECTIONS"))
		dbMinConns, _ := strconv.Atoi(os.Getenv("DB_MIN_CONNECTIONS"))
		dbMaxConnLifetime, _ := strconv.Atoi(os.Getenv("DB_MAX_CONN_LIFETIME"))
		dbMaxConnIdleTime, _ := strconv.Atoi(os.Getenv("DB_MAX_CONN_IDLE_TIME"))

		config = &Config{
			Server: ServerConfig{
				GoEnv:          os.Getenv("GO_ENV"),
				ServiceName:    os.Getenv("SERVICE_NAME"),
				ServiceVersion: os.Getenv("SERVICE_VERSION"),
				Host:           os.Getenv("SERVER_HOST"),
				Port:           os.Getenv("SERVER_PORT"),
			},
			JWT: JWTConfig{
				AccessTokenSecret:      os.Getenv("JWT_ACCESS_TOKEN_SECRET"),
				AccessTokenExpiration:  accessTokenExp,
				RefreshTokenSecret:     os.Getenv("JWT_REFRESH_TOKEN_SECRET"),
				RefreshTokenExpiration: refreshTokenExp,
			},
			RateLimit: RateLimitConfig{
				RPS:   rateRPS,
				Burst: rateBurst,
			},
			Redis: RedisConfig{
				Host:     os.Getenv("REDIS_HOST"),
				Port:     os.Getenv("REDIS_PORT"),
				Password: os.Getenv("REDIS_PASSWORD"),
				CacheDB:  redisCacheDB,
				QueueDB:  redisQueueDB,
				Prefix:   os.Getenv("REDIS_PREFIX"),
			},
			CORS: CORSConfig{
				AllowOrigins:     os.Getenv("CORS_ALLOW_ORIGINS"),
				AllowMethods:     os.Getenv("CORS_ALLOW_METHODS"),
				AllowHeaders:     os.Getenv("CORS_ALLOW_HEADERS"),
				AllowCredentials: corsAllowCredentials,
			},
			Database: DatabaseConfig{
				Host:            os.Getenv("DB_HOST"),
				Port:            dbPort,
				User:            os.Getenv("DB_USER"),
				Password:        os.Getenv("DB_PASSWORD"),
				DBName:          os.Getenv("DB_NAME"),
				Schema:          os.Getenv("DB_SCHEMA"),
				SSLMode:         os.Getenv("DB_SSL_MODE"),
				MaxConnections:  dbMaxConns,
				MinConnections:  dbMinConns,
				MaxConnLifetime: dbMaxConnLifetime,
				MaxConnIdleTime: dbMaxConnIdleTime,
			},
		}
	})

	return config
}

func (c *CORSConfig) GetAllowOrigins() string {
	if c.AllowOrigins == "" || c.AllowOrigins == "*" {
		return "*"
	}
	return c.AllowOrigins
}
