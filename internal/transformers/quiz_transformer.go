package transformers

import (
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
)

func ConvertToQuizModel(result interface{}) *models.Quiz {
	switch v := result.(type) {
	case sqlc.GetQuizWithOwnerRow:
		return convertToQuizWithOwnerModel(v)
	case sqlc.GetPublicQuizzesRow:
		return convertToQuizWithOwnerFromPublicQuizzes(v)
	case sqlc.Quiz:
		return convertToQuizModel(v)
	default:
		panic("unsupported quiz result type")
	}
}

func convertToQuizModel(result sqlc.Quiz) *models.Quiz {
	var publishedAt *time.Time
	if result.PublishedAt.Valid {
		publishedAt = &result.PublishedAt.Time
	}

	return &models.Quiz{
		ID:                   result.ID,
		Title:                result.Title,
		Description:          result.Description,
		Visibility:           models.QuizVisibility(result.Visibility),
		OwnerID:              result.OwnerID,
		MaxParticipants:      result.MaxParticipants,
		CurrentQuestionIndex: result.CurrentQuestionIndex,
		TotalQuestions:       result.TotalQuestions,
		CreatedAt:            result.CreatedAt.Time,
		UpdatedAt:            result.UpdatedAt.Time,
		PublishedAt:          publishedAt,
	}
}

func convertToQuizWithOwnerModel(result sqlc.GetQuizWithOwnerRow) *models.Quiz {
	var publishedAt *time.Time
	if result.PublishedAt.Valid {
		publishedAt = &result.PublishedAt.Time
	}

	quiz := &models.Quiz{
		ID:                   result.ID,
		Title:                result.Title,
		Description:          result.Description,
		Visibility:           models.QuizVisibility(result.Visibility),
		OwnerID:              result.OwnerID,
		MaxParticipants:      result.MaxParticipants,
		CurrentQuestionIndex: result.CurrentQuestionIndex,
		TotalQuestions:       result.TotalQuestions,
		CreatedAt:            result.CreatedAt.Time,
		UpdatedAt:            result.UpdatedAt.Time,
		PublishedAt:          publishedAt,
	}

	if result.OwnerUserID != nil {
		quiz.Owner = &models.Owner{
			ID:        *result.OwnerUserID,
			Username:  *result.OwnerUsername,
			Email:     *result.OwnerEmail,
			AvatarURL: result.OwnerAvatarUrl,
		}
	}

	return quiz
}

func convertToQuizWithOwnerFromPublicQuizzes(result sqlc.GetPublicQuizzesRow) *models.Quiz {
	var publishedAt *time.Time
	if result.PublishedAt.Valid {
		publishedAt = &result.PublishedAt.Time
	}

	quiz := &models.Quiz{
		ID:                   result.ID,
		Title:                result.Title,
		Description:          result.Description,
		Visibility:           models.QuizVisibility(result.Visibility),
		OwnerID:              result.OwnerID,
		MaxParticipants:      result.MaxParticipants,
		CurrentQuestionIndex: result.CurrentQuestionIndex,
		TotalQuestions:       result.TotalQuestions,
		CreatedAt:            result.CreatedAt.Time,
		UpdatedAt:            result.UpdatedAt.Time,
		PublishedAt:          publishedAt,
	}

	if result.OwnerUserID != nil {
		quiz.Owner = &models.Owner{
			ID:        *result.OwnerUserID,
			Username:  *result.OwnerUsername,
			Email:     *result.OwnerEmail,
			AvatarURL: result.OwnerAvatarUrl,
		}
	}

	return quiz
}
