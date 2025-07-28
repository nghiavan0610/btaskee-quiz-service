package repositories

import (
	"context"
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
)

type (
	UserRepository interface {
		GetUserDetail(ctx context.Context, id int64) (*models.User, error)
		FindValidateName(ctx context.Context, username string, excludeId *int64) (*models.User, error)
		UpdateUser(ctx context.Context, user *models.User) (*models.User, error)
		DeleteUser(ctx context.Context, id int64) error
		UpdateUserLastLogin(ctx context.Context, id int64) (*models.User, error)
	}

	userRepository struct {
		queries *sqlc.Queries
	}
)

func ProvideUserRepository(queries *sqlc.Queries) UserRepository {
	return &userRepository{
		queries: queries,
	}
}

func (r *userRepository) GetUserDetail(ctx context.Context, id int64) (*models.User, error) {
	result, err := r.queries.GetUserDetail(ctx, id)
	if err != nil {
		return nil, err
	}

	var lastLoginAt *time.Time
	if result.LastLoginAt.Valid {
		lastLoginAt = &result.LastLoginAt.Time
	}

	return &models.User{
		ID:          result.ID,
		Username:    result.Username,
		Email:       result.Email,
		IsActive:    result.IsActive,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
		LastLoginAt: lastLoginAt,
	}, nil
}

func (r *userRepository) FindValidateName(ctx context.Context, username string, excludeId *int64) (*models.User, error) {
	params := sqlc.FindValidateNameParams{
		Username:  username,
		ExcludeID: excludeId,
	}

	result, err := r.queries.FindValidateName(ctx, params)
	if err != nil {
		return nil, err
	}

	return &models.User{
		ID:       result.ID,
		Username: result.Username,
	}, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *models.User) (*models.User, error) {
	params := sqlc.UpdateUserParams{
		ID:       user.ID,
		Username: user.Username,
	}

	result, err := r.queries.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	var lastLoginAt *time.Time
	if result.LastLoginAt.Valid {
		lastLoginAt = &result.LastLoginAt.Time
	}

	return &models.User{
		ID:          result.ID,
		Username:    result.Username,
		Email:       result.Email,
		IsActive:    result.IsActive,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
		LastLoginAt: lastLoginAt,
	}, nil
}

func (r *userRepository) DeleteUser(ctx context.Context, id int64) error {
	return r.queries.DeleteUser(ctx, id)
}

func (r *userRepository) UpdateUserLastLogin(ctx context.Context, id int64) (*models.User, error) {
	result, err := r.queries.UpdateUserLastLogin(ctx, id)
	if err != nil {
		return nil, err
	}

	var lastLoginAt *time.Time
	if result.LastLoginAt.Valid {
		lastLoginAt = &result.LastLoginAt.Time
	}

	return &models.User{
		ID:          result.ID,
		Username:    result.Username,
		Email:       result.Email,
		IsActive:    result.IsActive,
		CreatedAt:   result.CreatedAt.Time,
		UpdatedAt:   result.UpdatedAt.Time,
		LastLoginAt: lastLoginAt,
	}, nil
}
