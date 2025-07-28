package repositories

import (
	"context"
	"fmt"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/transformers"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
)

type (
	QuestionRepository interface {
		CreateQuestion(ctx context.Context, question *models.Question) (*models.Question, error)
		GetQuestionByID(ctx context.Context, id int64) (*models.Question, error)
		GetMaxQuestionIndexByQuiz(ctx context.Context, id int64) (int32, error)
		GetQuestionListByQuiz(ctx context.Context, quizID int64) ([]*models.Question, error)
		CountQuestionsByQuiz(ctx context.Context, quizID int64) (int64, error)
		UpdateQuestion(ctx context.Context, question *models.Question) (*models.Question, error)
		UpdateQuestionIndex(ctx context.Context, questionID int64, index int32) error
		UpdateQuestionIndexesBatch(ctx context.Context, questionIDs []int64, indexes []int32) error
		DeleteQuestion(ctx context.Context, id int64) error
	}

	questionRepository struct {
		queries *sqlc.Queries
	}
)

func ProvideQuestionRepository(queries *sqlc.Queries) QuestionRepository {
	return &questionRepository{
		queries: queries,
	}
}

func (r *questionRepository) CreateQuestion(ctx context.Context, question *models.Question) (*models.Question, error) {
	answersBytes, err := transformers.ConvertAnswersToJSON(question.Answers)
	if err != nil {
		return nil, err
	}

	params := sqlc.CreateQuestionParams{
		QuizID:    question.QuizID,
		Question:  question.Question,
		Type:      sqlc.QuestionType(question.Type),
		Answers:   answersBytes,
		Index:     question.Index,
		TimeLimit: sqlc.TimeLimitType(question.TimeLimit),
	}

	result, err := r.queries.CreateQuestion(ctx, params)
	if err != nil {
		return nil, err
	}

	return transformers.ConvertToQuestionModel(result)
}

func (r *questionRepository) GetQuestionByID(ctx context.Context, id int64) (*models.Question, error) {
	result, err := r.queries.GetQuestionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return transformers.ConvertToQuestionModel(result)
}

func (r *questionRepository) GetMaxQuestionIndexByQuiz(ctx context.Context, id int64) (int32, error) {
	result, err := r.queries.GetMaxQuestionIndexByQuiz(ctx, id)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (r *questionRepository) GetQuestionListByQuiz(ctx context.Context, quizID int64) ([]*models.Question, error) {
	results, err := r.queries.GetQuestionListByQuiz(ctx, quizID)
	if err != nil {
		return nil, err
	}

	questions := make([]*models.Question, len(results))
	for i, result := range results {
		question, err := transformers.ConvertToQuestionModel(result)
		if err != nil {
			return nil, err
		}
		questions[i] = question
	}

	return questions, nil
}

func (r *questionRepository) CountQuestionsByQuiz(ctx context.Context, quizID int64) (int64, error) {
	return r.queries.CountQuestionsByQuiz(ctx, quizID)
}

func (r *questionRepository) UpdateQuestion(ctx context.Context, question *models.Question) (*models.Question, error) {
	answersBytes, err := transformers.ConvertAnswersToJSON(question.Answers)
	if err != nil {
		return nil, exception.InternalError("INVALID_ANSWERS_FORMAT", "Failed to process question answers").
			WithDetails(err.Error())
	}

	params := sqlc.UpdateQuestionParams{
		ID:        question.ID,
		Question:  question.Question,
		Type:      sqlc.QuestionType(question.Type),
		Answers:   answersBytes,
		TimeLimit: sqlc.TimeLimitType(question.TimeLimit),
	}

	result, err := r.queries.UpdateQuestion(ctx, params)
	if err != nil {
		return nil, err
	}

	return transformers.ConvertToQuestionModel(result)
}

func (r *questionRepository) DeleteQuestion(ctx context.Context, id int64) error {
	err := r.queries.DeleteQuestion(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *questionRepository) UpdateQuestionIndex(ctx context.Context, questionID int64, index int32) error {
	params := sqlc.UpdateQuestionIndexParams{
		ID:    questionID,
		Index: index,
	}

	err := r.queries.UpdateQuestionIndex(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (r *questionRepository) UpdateQuestionIndexesBatch(ctx context.Context, questionIDs []int64, indexes []int32) error {
	if len(questionIDs) != len(indexes) {
		return fmt.Errorf("questionIDs and indexes length mismatch")
	}

	// Sequential batch update
	for i, questionID := range questionIDs {
		params := sqlc.UpdateQuestionIndexParams{
			ID:    questionID,
			Index: indexes[i],
		}

		err := r.queries.UpdateQuestionIndex(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update question %d: %w", questionID, err)
		}
	}

	return nil
}
