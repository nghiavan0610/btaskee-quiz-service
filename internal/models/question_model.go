package models

import (
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
)

type QuestionType = sqlc.QuestionType

const (
	QuestionTypeSingleChoice   = sqlc.QuestionTypeSingleChoice
	QuestionTypeMultipleChoice = sqlc.QuestionTypeMultipleChoice
	QuestionTypeTextInput      = sqlc.QuestionTypeTextInput
)

type TimeLimitType = sqlc.TimeLimitType

const (
	TimeLimitType5  = sqlc.TimeLimitType5  // 5 seconds
	TimeLimitType10 = sqlc.TimeLimitType10 // 10 seconds
	TimeLimitType20 = sqlc.TimeLimitType20 // 20 seconds
	TimeLimitType45 = sqlc.TimeLimitType45 // 45 seconds
	TimeLimitType80 = sqlc.TimeLimitType80 // 80 seconds
)

type AnswerData struct {
	Text      string `json:"text"`
	IsCorrect bool   `json:"is_correct"`
}

type Question struct {
	ID        int64         `json:"id"`
	QuizID    int64         `json:"quiz_id"`
	Question  string        `json:"question"`
	Type      QuestionType  `json:"type"`
	Answers   []AnswerData  `json:"answers"`
	TimeLimit TimeLimitType `json:"time_limit"`
	Index     int32         `json:"index"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
