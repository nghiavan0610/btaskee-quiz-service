package services

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
)

type (
	ValidationService interface {
		ValidateQuizOwnership(ctx context.Context, quizID int64, userID int64, includeQuestions bool) (*models.Quiz, *exception.AppError)
		ValidateQuestionOwnership(ctx context.Context, questionID int64, userID int64) (*models.Question, *models.Quiz, *exception.AppError)
	}

	validationService struct {
		quizRepo     repositories.QuizRepository
		questionRepo repositories.QuestionRepository
	}
)

var (
	validationServiceOnce     sync.Once
	validationServiceInstance ValidationService
)

func ProvideValidationService(
	quizRepo repositories.QuizRepository,
	questionRepo repositories.QuestionRepository,
) ValidationService {
	validationServiceOnce.Do(func() {
		validationServiceInstance = &validationService{
			quizRepo:     quizRepo,
			questionRepo: questionRepo,
		}
	})
	return validationServiceInstance
}

func (s *validationService) ValidateQuizOwnership(ctx context.Context, quizID int64, userID int64, includeQuestions bool) (*models.Quiz, *exception.AppError) {
	quiz, err := s.quizRepo.GetQuizDetail(ctx, quizID, includeQuestions)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrQuizNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	if quiz.OwnerID != userID {
		return nil, exception.Forbidden(errors.CodeUnauthorized, errors.ErrUnauthorizedAccess).
			WithDetails("Only the quiz owner can perform this action")
	}

	return quiz, nil
}

func (s *validationService) ValidateQuestionOwnership(ctx context.Context, questionID int64, userID int64) (*models.Question, *models.Quiz, *exception.AppError) {
	question, err := s.questionRepo.GetQuestionByID(ctx, questionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, exception.NotFound(errors.CodeNotFound, errors.ErrQuestionNotFound)
		}
		return nil, nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	quiz, appErr := s.ValidateQuizOwnership(ctx, question.QuizID, userID, false)
	if appErr != nil {
		return nil, nil, appErr
	}

	return question, quiz, nil
}
