package dtos

import (
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
)

type GetMyQuizDetailRequest struct {
	QuizID int64 `params:"quiz_id" validate:"required"`
}

type CreateQuizRequest struct {
	Title           string  `json:"title" validate:"required,min=3,max=255"`
	Description     *string `json:"description" validate:"omitempty,max=1000"`
	MaxParticipants int32   `json:"max_participants" validate:"min=1,max=1000"`
}

type UpdateQuizRequest struct {
	QuizID          int64                 `params:"quiz_id" validate:"required"`
	Title           string                `json:"title" validate:"required,min=3,max=255"`
	Description     *string               `json:"description" validate:"omitempty,max=1000"`
	MaxParticipants int32                 `json:"max_participants" validate:"min=1,max=1000"`
	Visibility      models.QuizVisibility `json:"visibility" validate:"required,oneof=private unlisted published"`
}

type GetMyQuizListRequest struct {
	Query      *string                `query:"query"`
	Visibility *models.QuizVisibility `query:"visibility" validate:"omitempty,oneof='' private unlisted published"`
	Page       int32                  `query:"page" validate:"min=1"`
	Limit      int32                  `query:"limit" validate:"min=1,max=100"`
}

type DeleteQuizRequest struct {
	QuizID int64 `params:"quiz_id" validate:"required"`
}

type GetQuizListRequest struct {
	Query  *string `query:"query"`
	Cursor *string `query:"cursor"`
	Limit  int32   `query:"limit" validate:"min=1,max=100"`
	SortBy *string `query:"sort_by" validate:"omitempty,oneof='' name_asc name_desc time_newest time_oldest view_count play_count"`
}

type GetQuizListResponse struct {
	Quizzes    []*models.Quiz `json:"quizzes"`
	NextCursor *string        `json:"next_cursor,omitempty"`
	HasMore    bool           `json:"has_more"`
}

type GetQuizDetailRequest struct {
	QuizID int64 `params:"quiz_id" validate:"required"`
}
