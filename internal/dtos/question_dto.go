package dtos

import "github.com/nghiavan0610/btaskee-quiz-service/internal/models"

type CreateQuestionRequest struct {
	QuizID    int64                `json:"quiz_id" validate:"required"`
	Question  string               `json:"question" validate:"required,min=1,max=100"`
	Type      models.QuestionType  `json:"type" validate:"required,oneof=single_choice multiple_choice_checkbox text_input"`
	Answers   []models.AnswerData  `json:"answers"`
	TimeLimit models.TimeLimitType `json:"time_limit" validate:"oneof=5 10 20 45 80"` // Predefined time limits
}

type UpdateQuestionRequest struct {
	QuestionID int64                `params:"question_id" validate:"required"`
	Question   string               `json:"question" validate:"required,min=1,max=100"`
	Type       models.QuestionType  `json:"type" validate:"required,oneof=single_choice multiple_choice_checkbox text_input"`
	Answers    []models.AnswerData  `json:"answers"`
	TimeLimit  models.TimeLimitType `json:"time_limit" validate:"oneof=5 10 20 45 80"` // Predefined time limits
}

type QuestionIndexesPayload struct {
	QuestionID int64 `json:"question_id" validate:"required"`
	Index      int32 `json:"index" validate:"required,min=0"`
}

type UpdateQuestionIndexRequest struct {
	Indexes []QuestionIndexesPayload `json:"indexes" validate:"required,dive"`
}

type DeleteQuestionRequest struct {
	QuestionID int64 `params:"question_id" validate:"required"`
}
