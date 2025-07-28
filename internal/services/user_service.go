package services

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
)

type (
	UserService interface {
		GetMine(ctx context.Context, authUser *dtos.UserSession) (*models.User, *exception.AppError)
		UpdateMine(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateMineRequest) (*models.User, *exception.AppError)
	}

	userService struct {
		userRepo     repositories.UserRepository
		tokenService TokenService
		logger       *logger.Logger
	}
)

var (
	userServiceOnce     sync.Once
	userServiceInstance UserService
)

func ProvideUserService(
	userRepo repositories.UserRepository,
	tokenService TokenService,
	logger *logger.Logger,
) UserService {
	userServiceOnce.Do(func() {
		userServiceInstance = &userService{
			userRepo:     userRepo,
			tokenService: tokenService,
			logger:       logger,
		}
	})
	return userServiceInstance
}

func (s *userService) GetMine(ctx context.Context, authUser *dtos.UserSession) (*models.User, *exception.AppError) {
	s.logger.Info("[GET MINE]", authUser)

	user, err := s.userRepo.GetUserDetail(ctx, authUser.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrUserNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	return user, nil
}

func (s *userService) UpdateMine(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateMineRequest) (*models.User, *exception.AppError) {
	s.logger.Info("[UPDATE MINE]", authUser, req)

	existingName, err := s.userRepo.FindValidateName(ctx, req.Username, &authUser.UserID)
	if err != nil && err != pgx.ErrNoRows {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}
	if existingName != nil {
		return nil, exception.Conflict(errors.CodeAlreadyExists, errors.ErrUsernameAlreadyExists)
	}

	user, err := s.userRepo.UpdateUser(ctx, &models.User{
		ID:       authUser.UserID,
		Username: req.Username,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrUserNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	return user, nil
}
