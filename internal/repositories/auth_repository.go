package repositories

import (
	"context"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
)

type (
	AuthRepository interface {
		CheckEmailOrUsernameExists(ctx context.Context, email, username string) (*dtos.SignUpConflictResult, error)
		RegisterAccount(ctx context.Context, user *models.User) (*models.User, error)
		GetUserByEmailIncludePassword(ctx context.Context, email string) (*models.User, error)
	}

	authRepository struct {
		queries *sqlc.Queries
	}
)

func ProvideAuthRepository(queries *sqlc.Queries) AuthRepository {
	return &authRepository{
		queries: queries,
	}
}

func (r *authRepository) CheckEmailOrUsernameExists(ctx context.Context, email, username string) (*dtos.SignUpConflictResult, error) {
	params := sqlc.CheckEmailOrUsernameExistsParams{
		Email:    email,
		Username: username,
	}

	result, err := r.queries.CheckEmailOrUsernameExists(ctx, params)
	if err != nil {
		return nil, err
	}

	conflictField := ""
	if result.ConflictField != nil {
		if cf, ok := result.ConflictField.(string); ok {
			conflictField = cf
		}
	}

	return &dtos.SignUpConflictResult{
		ID:            result.ID,
		ConflictField: conflictField,
	}, nil
}

func (r *authRepository) RegisterAccount(ctx context.Context, user *models.User) (*models.User, error) {
	params := sqlc.RegisterAccountParams{
		Username: user.Username,
		Email:    user.Email,
		Password: user.Password,
	}

	result, err := r.queries.RegisterAccount(ctx, params)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        result.ID,
		Username:  result.Username,
		Email:     result.Email,
		IsActive:  result.IsActive,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}

func (r *authRepository) GetUserByEmailIncludePassword(ctx context.Context, email string) (*models.User, error) {
	result, err := r.queries.GetUserByEmailIncludePassword(ctx, email)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:        result.ID,
		Username:  result.Username,
		Email:     result.Email,
		Password:  result.Password,
		IsActive:  result.IsActive,
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}, nil
}
