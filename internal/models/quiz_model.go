package models

import (
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/response"
)

type QuizVisibility = sqlc.QuizVisibility

const (
	QuizVisibilityPrivate   = sqlc.QuizVisibilityPrivate
	QuizVisibilityPublished = sqlc.QuizVisibilityPublished
)

type Owner struct {
	ID        int64   `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	AvatarURL *string `json:"avatar_url"`
}

type Quiz struct {
	ID                   int64          `json:"id"`
	Title                string         `json:"title"`
	Description          *string        `json:"description,omitempty"`
	OwnerID              int64          `json:"owner_id"`
	Visibility           QuizVisibility `json:"visibility"`
	Slug                 *string        `json:"slug,omitempty"`
	ViewCount            int32          `json:"view_count"`
	PlayCount            int32          `json:"play_count"`
	MaxParticipants      *int32         `json:"max_participants,omitempty"`
	CurrentQuestionIndex int32          `json:"current_question_index"`
	TotalQuestions       *int32         `json:"total_questions"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	PublishedAt          *time.Time     `json:"published_at,omitempty"`
	Questions            []Question     `json:"questions,omitempty"`
	Owner                *Owner         `json:"owner,omitempty"`
}

type MyQuizListResponse struct {
	Quizzes    []*Quiz                   `json:"quizzes"`
	Pagination response.OffsetPagination `json:"pagination"`
}
