package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/cache"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
)

type (
	TokenService interface {
		GenerateTokenPair(ctx context.Context, user *models.User) (*dtos.TokenResponse, *exception.AppError)
		ValidateAccessToken(ctx context.Context, tokenString string) (*dtos.JWTClaims, *exception.AppError)
		ValidateRefreshToken(ctx context.Context, tokenString string) (*dtos.JWTClaims, *exception.AppError)
		RevokeToken(ctx context.Context, userID string)
	}

	tokenService struct {
		cache           cache.Cache
		config          *config.Config
		accessTokenTTL  time.Duration
		refreshTokenTTL time.Duration
	}
)

var (
	tokenServiceOnce     sync.Once
	tokenServiceInstance TokenService
)

func ProvideTokenService(
	cache cache.Cache,
	config *config.Config,
) TokenService {
	tokenServiceOnce.Do(func() {
		tokenServiceInstance = &tokenService{
			cache:           cache,
			config:          config,
			accessTokenTTL:  time.Duration(config.JWT.AccessTokenExpiration) * time.Second,
			refreshTokenTTL: time.Duration(config.JWT.RefreshTokenExpiration) * time.Second,
		}
	})
	return tokenServiceInstance
}

func (s *tokenService) GenerateTokenPair(ctx context.Context, user *models.User) (*dtos.TokenResponse, *exception.AppError) {
	now := time.Now()

	userSession := dtos.UserSession{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		IsActive: user.IsActive,
	}

	// Generate access token
	accessClaims := &dtos.JWTClaims{
		UserSession: userSession,
		Type:        string(constants.CachePrefixAccessToken),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Server.ServiceName,
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        fmt.Sprintf("access_%d_%d", user.ID, now.Unix()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.config.JWT.AccessTokenSecret))
	if err != nil {
		return nil, exception.InternalError(errors.CodeTokenGenerateFailed, errors.ErrTokenGenerateFailed).
			WithDetails("Error occurred while generating access token").
			WithMetadata("error", err.Error())
	}

	// Generate refresh token
	refreshClaims := &dtos.JWTClaims{
		UserSession: userSession,
		Type:        string(constants.CachePrefixRefreshToken),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.Server.ServiceName,
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        fmt.Sprintf("refresh_%d_%d", user.ID, now.Unix()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.config.JWT.RefreshTokenSecret))
	if err != nil {
		return nil, exception.InternalError(errors.CodeTokenGenerateFailed, errors.ErrTokenGenerateFailed).
			WithDetails("Error occurred while generating refresh token").
			WithMetadata("error", err.Error())
	}

	// Cache both tokens
	if err := s.cacheToken(ctx, userSession, string(constants.CachePrefixAccessToken), now.Add(s.accessTokenTTL)); err != nil {
		return nil, err
	}

	if err := s.cacheToken(ctx, userSession, string(constants.CachePrefixRefreshToken), now.Add(s.refreshTokenTTL)); err != nil {
		return nil, err
	}

	return &dtos.TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int(s.accessTokenTTL.Seconds()),
	}, nil
}

func (s *tokenService) ValidateAccessToken(ctx context.Context, tokenString string) (*dtos.JWTClaims, *exception.AppError) {
	return s.validateToken(ctx, tokenString, s.config.JWT.AccessTokenSecret, string(constants.CachePrefixAccessToken))
}

func (s *tokenService) ValidateRefreshToken(ctx context.Context, tokenString string) (*dtos.JWTClaims, *exception.AppError) {
	return s.validateToken(ctx, tokenString, s.config.JWT.RefreshTokenSecret, string(constants.CachePrefixRefreshToken))
}

func (s *tokenService) validateToken(ctx context.Context, tokenString, secret, expectedType string) (*dtos.JWTClaims, *exception.AppError) {
	// Parse and validate JWT
	token, err := jwt.ParseWithClaims(tokenString, &dtos.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, exception.Unauthorized(errors.CodeTokenInvalid, errors.ErrTokenInvalid).
			WithDetails("Token is malformed or expired").
			WithMetadata("error", err.Error())
	}

	claims, ok := token.Claims.(*dtos.JWTClaims)
	if !ok || !token.Valid {
		return nil, exception.Unauthorized(errors.CodeTokenInvalid, errors.ErrTokenInvalid).
			WithDetails("Token claims are invalid")
	}

	if claims.Type != expectedType {
		return nil, exception.Unauthorized(errors.CodeTokenInvalid, errors.ErrTokenInvalid).
			WithDetails(fmt.Sprintf("Expected %s token but got %s", expectedType, claims.Type))
	}

	return claims, nil
}

func (s *tokenService) RevokeToken(ctx context.Context, userID string) {
	accessOpt := cache.CacheKeyOption{
		Module:    string(constants.CacheModuleAuth),
		Prefix:    string(constants.CachePrefixAccessToken),
		UniqueKey: userID,
	}

	refreshOpt := cache.CacheKeyOption{
		Module:    string(constants.CacheModuleAuth),
		Prefix:    string(constants.CachePrefixRefreshToken),
		UniqueKey: userID,
	}

	s.cache.Del(ctx, accessOpt)
	s.cache.Del(ctx, refreshOpt)
}

func (s *tokenService) cacheToken(ctx context.Context, userSession dtos.UserSession, tokenType string, expiresAt time.Time) *exception.AppError {
	cacheOpt := cache.CacheKeyOption{
		Module:    string(constants.CacheModuleAuth),
		Prefix:    tokenType,
		UniqueKey: fmt.Sprintf("%d", userSession.UserID),
		Value:     userSession,
		TTL:       time.Until(expiresAt),
	}

	if err := s.cache.Set(ctx, cacheOpt); err != nil {
		return exception.InternalError(errors.CodeCacheSetFailed, errors.ErrFailedToSetCache).
			WithDetails("Error occurred while caching token").
			WithMetadata("error", err.Error())
	}

	return nil
}
