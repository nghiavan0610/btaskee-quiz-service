package services

import (
	"context"
	"sync"
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/transformers"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
)

type (
	QuestionService interface {
		CreateQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.CreateQuestionRequest) (*models.Question, *exception.AppError)
		UpdateQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateQuestionRequest) (*models.Question, *exception.AppError)
		UpdateQuestionIndex(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateQuestionIndexRequest) *exception.AppError
		DeleteQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.DeleteQuestionRequest) *exception.AppError
	}

	questionService struct {
		logger            *logger.Logger
		questionRepo      repositories.QuestionRepository
		quizRepo          repositories.QuizRepository
		validationService ValidationService
	}
)

var (
	questionServiceOnce     sync.Once
	questionServiceInstance QuestionService
)

func ProvideQuestionService(
	logger *logger.Logger,
	questionRepo repositories.QuestionRepository,
	quizRepo repositories.QuizRepository,
	validationService ValidationService,
) QuestionService {
	questionServiceOnce.Do(func() {
		questionServiceInstance = &questionService{
			logger:            logger,
			questionRepo:      questionRepo,
			quizRepo:          quizRepo,
			validationService: validationService,
		}
	})
	return questionServiceInstance
}

func (s *questionService) CreateQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.CreateQuestionRequest) (*models.Question, *exception.AppError) {
	s.logger.Info("[CREATE QUESTION]", authUser, req)

	_, appErr := s.validationService.ValidateQuizOwnership(ctx, req.QuizID, authUser.UserID, false)
	if appErr != nil {
		return nil, appErr
	}

	// Validate question answers based on type
	if err := transformers.ValidateAnswersFormat(req.Answers, req.Type); err != nil {
		return nil, exception.BadRequest(errors.CodeBadRequest, err.Error()).
			WithDetails("Invalid question answers format")
	}

	// Get next order index
	lastIndex, err := s.questionRepo.GetMaxQuestionIndexByQuiz(ctx, req.QuizID)
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	question := &models.Question{
		QuizID:    req.QuizID,
		Question:  req.Question,
		Index:     lastIndex + 1,
		Type:      models.QuestionType(req.Type),
		Answers:   req.Answers,
		TimeLimit: models.TimeLimitType(req.TimeLimit),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdQuestion, err := s.questionRepo.CreateQuestion(ctx, question)
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	// Update the quiz's total_questions count
	s.updateQuizTotalQuestions(ctx, req.QuizID)

	return createdQuestion, nil
}

func (s *questionService) UpdateQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateQuestionRequest) (*models.Question, *exception.AppError) {
	s.logger.Info("[UPDATE QUESTION]", authUser, req)

	s.validationService.ValidateQuestionOwnership(ctx, req.QuestionID, authUser.UserID)

	// Validate question answers based on type
	if err := transformers.ValidateAnswersFormat(req.Answers, req.Type); err != nil {
		return nil, exception.BadRequest(errors.CodeBadRequest, err.Error()).
			WithDetails("Invalid question answers format")
	}

	question := &models.Question{
		ID:        req.QuestionID,
		Question:  req.Question,
		Type:      models.QuestionType(req.Type),
		Answers:   req.Answers,
		TimeLimit: models.TimeLimitType(req.TimeLimit),
		UpdatedAt: time.Now(),
	}

	updatedQuestion, err := s.questionRepo.UpdateQuestion(ctx, question)
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	return updatedQuestion, nil
}

func (s *questionService) UpdateQuestionIndex(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateQuestionIndexRequest) *exception.AppError {
	s.logger.Info("[UPDATE QUESTION INDEX]", authUser, req)

	if len(req.Indexes) == 0 {
		return exception.BadRequest(errors.CodeBadRequest, "No question indexes to update")
	}

	_, _, appErr := s.validationService.ValidateQuestionOwnership(ctx, req.Indexes[0].QuestionID, authUser.UserID)
	if appErr != nil {
		return appErr
	}

	questionIDs := make([]int64, len(req.Indexes))
	indexes := make([]int32, len(req.Indexes))

	for i, indexPayload := range req.Indexes {
		questionIDs[i] = indexPayload.QuestionID
		indexes[i] = indexPayload.Index
	}

	if err := s.questionRepo.UpdateQuestionIndexesBatch(ctx, questionIDs, indexes); err != nil {
		return exception.InternalError(errors.CodeDBError, err.Error()).
			WithDetails("Failed to update question indexes")
	}

	return nil
}

func (s *questionService) DeleteQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.DeleteQuestionRequest) *exception.AppError {
	s.logger.Info("[DELETE QUESTION]", authUser, req)

	question, _, appErr := s.validationService.ValidateQuestionOwnership(ctx, req.QuestionID, authUser.UserID)
	if appErr != nil {
		return appErr
	}

	err := s.questionRepo.DeleteQuestion(ctx, req.QuestionID)
	if err != nil {
		return exception.InternalError(errors.CodeDBError, err.Error())
	}

	// Update the quiz's total_questions count
	s.updateQuizTotalQuestions(ctx, question.QuizID)

	return nil
}

func (s *questionService) updateQuizTotalQuestions(ctx context.Context, quizID int64) error {
	totalQuestions, err := s.questionRepo.CountQuestionsByQuiz(ctx, quizID)
	if err != nil {
		return err
	}

	return s.quizRepo.UpdateTotalQuestions(ctx, quizID, int32(totalQuestions))
}
