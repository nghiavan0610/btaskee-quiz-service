package services

import (
	"context"
	goErrors "errors"
	"strconv"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/utils"
)

type (
	AuthService interface {
		SignUp(ctx context.Context, req *dtos.SignUpRequest) (*dtos.TokenResponse, *exception.AppError)
		SignIn(ctx context.Context, req *dtos.SignInRequest) (*dtos.TokenResponse, *exception.AppError)
		RefreshToken(ctx context.Context, authUser *dtos.UserSession) (*dtos.TokenResponse, *exception.AppError)
		SignOut(ctx context.Context, authUser *dtos.UserSession)
	}

	authService struct {
		authRepo     repositories.AuthRepository
		tokenService TokenService
		logger       *logger.Logger
	}
)

var (
	authServiceOnce     sync.Once
	authServiceInstance AuthService
)

func ProvideAuthService(
	authRepo repositories.AuthRepository,
	tokenService TokenService,
	logger *logger.Logger,
) AuthService {
	authServiceOnce.Do(func() {
		authServiceInstance = &authService{
			authRepo:     authRepo,
			tokenService: tokenService,
			logger:       logger,
		}
	})
	return authServiceInstance
}

func (s *authService) SignUp(ctx context.Context, req *dtos.SignUpRequest) (*dtos.TokenResponse, *exception.AppError) {
	s.logger.Info("[SIGN UP]", req)

	conflict, err := s.authRepo.CheckEmailOrUsernameExists(ctx, req.Email, req.Username)
	if err == nil {
		switch conflict.ConflictField {
		case "email":
			return nil, exception.Conflict(errors.CodeAlreadyExists, errors.ErrEmailAlreadyExists)
		case "username":
			return nil, exception.Conflict(errors.CodeAlreadyExists, errors.ErrUsernameAlreadyExists)
		}
	}
	if !goErrors.Is(err, pgx.ErrNoRows) {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	// Hash password
	hashedPassword, err := utils.EncryptToHash(req.Password)
	if err != nil {
		return nil, exception.InternalError(errors.CodeInternal, errors.ErrPasswordHash)
	}

	// Create user
	createdUser, err := s.authRepo.RegisterAccount(ctx, &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		IsActive: true,
	})
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	// Generate tokens
	tokens, appErr := s.tokenService.GenerateTokenPair(ctx, createdUser)
	if appErr != nil {
		return nil, appErr
	}

	return &dtos.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *authService) SignIn(ctx context.Context, req *dtos.SignInRequest) (*dtos.TokenResponse, *exception.AppError) {
	s.logger.Info("[SIGN IN]", req)

	user, err := s.authRepo.GetUserByEmailIncludePassword(ctx, req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrUserNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	if !user.IsActive {
		return nil, exception.Forbidden(errors.CodeUnauthorized, errors.ErrUserDisabled).
			WithDetails("Your account has been disabled. Please contact support")
	}

	// Verify password
	if !utils.VerifyEncryptHash(user.Password, req.Password) {
		return nil, exception.Unauthorized(errors.CodeUnauthorized, errors.ErrInvalidCredentials).
			WithDetails("Email or password is incorrect")
	}

	// Update last login

	// Generate tokens
	tokens, appErr := s.tokenService.GenerateTokenPair(ctx, user)
	if appErr != nil {
		return nil, appErr
	}

	return &dtos.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, authUser *dtos.UserSession) (*dtos.TokenResponse, *exception.AppError) {
	s.logger.Info("[REFRESH TOKEN]", authUser)

	s.tokenService.RevokeToken(ctx, strconv.FormatInt(authUser.UserID, 10))

	// Generate new tokens
	tokens, appErr := s.tokenService.GenerateTokenPair(ctx, &models.User{
		ID:       authUser.UserID,
		Username: authUser.Username,
		Email:    authUser.Email,
		IsActive: authUser.IsActive,
	})
	if appErr != nil {
		return nil, appErr
	}

	return &dtos.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    tokens.ExpiresIn,
	}, nil
}

func (s *authService) SignOut(ctx context.Context, authUser *dtos.UserSession) {
	s.logger.Info("[SIGN OUT]", authUser)

	s.tokenService.RevokeToken(ctx, strconv.FormatInt(authUser.UserID, 10))
}
