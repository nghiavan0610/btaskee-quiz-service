package guards

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/services"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/cache"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
)

// guardConfig defines the configuration for a specific guard type
type guardConfig struct {
	TokenName    string
	ValidateFunc func(ctx context.Context, token string) (*dtos.JWTClaims, *exception.AppError)
	CachePrefix  string
	CheckActive  bool
}

type AuthGuard interface {
	AccessTokenGuard() fiber.Handler
	RefreshTokenGuard() fiber.Handler
	OptionalAccessTokenGuard() fiber.Handler
	GetAuthUser(c *fiber.Ctx) (*dtos.UserSession, *exception.AppError)
}

type authGuard struct {
	tokenService services.TokenService
	cache        cache.Cache
}

var (
	authGuardInstance AuthGuard
	authGuardOnce     sync.Once
)

func ProvideAuthGuard(tokenService services.TokenService, cache cache.Cache) AuthGuard {
	authGuardOnce.Do(func() {
		authGuardInstance = &authGuard{
			tokenService: tokenService,
			cache:        cache,
		}
	})
	return authGuardInstance
}

func (g *authGuard) AccessTokenGuard() fiber.Handler {
	return g.createGuard(guardConfig{
		TokenName:    "Access token",
		ValidateFunc: g.tokenService.ValidateAccessToken,
		CachePrefix:  string(constants.CachePrefixAccessToken),
		CheckActive:  true,
	})
}

func (g *authGuard) RefreshTokenGuard() fiber.Handler {
	return g.createGuard(guardConfig{
		TokenName:    "Refresh token",
		ValidateFunc: g.tokenService.ValidateRefreshToken,
		CachePrefix:  string(constants.CachePrefixRefreshToken),
		CheckActive:  true,
	})
}

func (g *authGuard) OptionalAccessTokenGuard() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token provided - continue without setting user (for anonymous access)
			return c.Next()
		}

		// Token provided - try to validate it
		if !strings.HasPrefix(authHeader, "Bearer ") {
			// Invalid token format - continue without setting user
			return c.Next()
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, appErr := g.tokenService.ValidateAccessToken(c.Context(), tokenString)
		if appErr != nil {
			// Invalid token - continue without setting user
			return c.Next()
		}

		// Check if user exists in claims
		if claims.UserSession.UserID == 0 {
			// Invalid user session - continue without setting user
			return c.Next()
		}

		// Check if token is in cache (active)
		var cachedUserSession dtos.UserSession
		cacheErr := g.cache.GetObject(context.Background(), cache.CacheKeyOption{
			Module:    string(constants.CacheModuleAuth),
			Prefix:    string(constants.CachePrefixAccessToken),
			UniqueKey: strconv.FormatInt(claims.UserSession.UserID, 10),
		}, &cachedUserSession)
		if cacheErr != nil {
			// Token not in cache or cache error - continue without setting user
			return c.Next()
		}

		// Check if user is active
		if !cachedUserSession.IsActive {
			// User not active - continue without setting user
			return c.Next()
		}

		// Valid token and active user - set user in locals
		c.Locals(string(constants.KEY_AUTH_USER), &cachedUserSession)

		return c.Next()
	}
}

func (g *authGuard) GetAuthUser(c *fiber.Ctx) (*dtos.UserSession, *exception.AppError) {
	userSession, ok := c.Locals(string(constants.KEY_AUTH_USER)).(*dtos.UserSession)
	if !ok {
		return nil, exception.Unauthorized(errors.CodeUnauthorized, errors.ErrUnauthorizedAccess).
			WithDetails("User session is missing")
	}
	return userSession, nil
}

func (g *authGuard) createGuard(config guardConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return exception.Unauthorized(errors.CodeUnauthorized, errors.ErrUnauthorizedAccess).
				WithDetails(config.TokenName + " required")
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			return exception.Unauthorized(errors.CodeUnauthorized, errors.ErrUnauthorizedAccess).
				WithDetails("Invalid authorization format")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, appErr := config.ValidateFunc(c.Context(), tokenString)
		if appErr != nil {
			return appErr
		}

		// Check if user exists in claims
		if claims.UserSession.UserID == 0 {
			return exception.Unauthorized(errors.CodeUnauthorized, errors.ErrUnauthorizedAccess).
				WithDetails("User session is missing from token claims")
		}

		userSession := claims.UserSession
		if config.CachePrefix != "" {
			var cachedUserSession dtos.UserSession
			appErr = g.cache.GetObject(context.Background(), cache.CacheKeyOption{
				Module:    string(constants.CacheModuleAuth),
				Prefix:    config.CachePrefix,
				UniqueKey: strconv.FormatInt(claims.UserSession.UserID, 10),
			}, &cachedUserSession)
			if appErr != nil {
				return exception.Unauthorized(errors.CodeUnauthorized, errors.ErrUnauthorizedAccess).
					WithDetails("User session is missing")
			}

			// Check if user is active (if required)
			if config.CheckActive && !cachedUserSession.IsActive {
				return exception.Unauthorized(errors.CodeUnauthorized, errors.ErrUserDisabled).
					WithDetails("User account is not active")
			}

			userSession = cachedUserSession
		}

		c.Locals(string(constants.KEY_AUTH_USER), &userSession)

		return c.Next()
	}
}
