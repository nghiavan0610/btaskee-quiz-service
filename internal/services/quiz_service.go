package services

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	helpers "github.com/nghiavan0610/btaskee-quiz-service/helpers/quiz"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/response"
	"github.com/nghiavan0610/btaskee-quiz-service/utils"
)

type (
	QuizService interface {
		// Quiz management - now with proper auth user context
		CreateQuiz(ctx context.Context, authUser *dtos.UserSession, req *dtos.CreateQuizRequest) (*models.Quiz, *exception.AppError)
		UpdateQuiz(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateQuizRequest) (*models.Quiz, *exception.AppError)
		DeleteQuiz(ctx context.Context, authUser *dtos.UserSession, req *dtos.DeleteQuizRequest) *exception.AppError
		GetMyQuizDetail(ctx context.Context, authUser *dtos.UserSession, req *dtos.GetMyQuizDetailRequest) (*models.Quiz, *exception.AppError)
		GetMyQuizList(ctx context.Context, authUser *dtos.UserSession, req *dtos.GetMyQuizListRequest) (*models.MyQuizListResponse, *exception.AppError)
		GetQuizList(ctx context.Context, req *dtos.GetQuizListRequest) (*dtos.GetQuizListResponse, *exception.AppError)
		GetQuizDetail(ctx context.Context, authUser *dtos.UserSession, req *dtos.GetQuizDetailRequest) (*models.Quiz, *exception.AppError)

		// GetUserQuizzes(ctx context.Context, authUser *dtos.UserSession) ([]*models.Quiz, *exception.AppError)

		// // Question management for quiz creation
		// AddQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.CreateQuestionRequest) (*models.Question, *exception.AppError)
		// UpdateQuestion(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateQuestionRequest) (*models.Question, *exception.AppError)
		// DeleteQuestion(ctx context.Context, authUser *dtos.UserSession, quizID, questionID int64) *exception.AppError
		// GetQuestions(ctx context.Context, authUser *dtos.UserSession, quizID int64) ([]*models.Question, *exception.AppError)

		// // Quiz lifecycle
		// PublishQuiz(ctx context.Context, authUser *dtos.UserSession, quizID int64) (*models.Quiz, *exception.AppError)
		// UnpublishQuiz(ctx context.Context, authUser *dtos.UserSession, quizID int64) (*models.Quiz, *exception.AppError)

		// // Join code management
		// GetQuizByJoinCode(ctx context.Context, joinCode string) (*models.Quiz, *exception.AppError)
		// GetQuizList(ctx context.Context) ([]*models.Quiz, *exception.AppError)

		// // Host controls (Kahoot-style)
		// StartQuiz(ctx context.Context, req *dtos.StartQuizRequest) (*models.Quiz, *exception.AppError)
		// EndQuiz(ctx context.Context, req *dtos.EndQuizRequest) (*models.Quiz, *exception.AppError)
		// NextQuestion(ctx context.Context, req *dtos.NextQuestionRequest) (*models.Quiz, *exception.AppError)

		// // Participant management
		// JoinQuizByCode(ctx context.Context, req *dtos.JoinQuizByCodeRequest) (*models.Participant, *exception.AppError)
		// JoinQuiz(ctx context.Context, req *dtos.JoinQuizRequest) (*models.Participant, *exception.AppError) // Legacy
		// SubmitAnswer(ctx context.Context, quizID, participantID, questionID int64, selectedAnswer string) (*models.Answer, *exception.AppError)

		// // Leaderboard & Status
		// GetLeaderboard(ctx context.Context, quizID int64) (*models.Leaderboard, *exception.AppError)
		// GetQuizStatus(ctx context.Context, quizID int64) (*models.QuizStatusInfo, *exception.AppError)
		// GetCurrentQuestion(ctx context.Context, quizID int64) (*models.Question, *exception.AppError)
		// GetLobbyInfo(ctx context.Context, quizID int64) (*models.LobbyInfo, *exception.AppError)
	}

	quizService struct {
		logger            *logger.Logger
		quizRepo          repositories.QuizRepository
		questionRepo      repositories.QuestionRepository
		validationService ValidationService
	}
)

type QuizSession struct {
	Quiz              *models.Quiz
	Questions         []models.Question
	CurrentQuestion   *models.Question
	StartTime         time.Time
	QuestionStartTime time.Time
}

var (
	quizServiceOnce     sync.Once
	quizServiceInstance QuizService
)

func ProvideQuizService(
	logger *logger.Logger,
	quizRepo repositories.QuizRepository,
	questionRepo repositories.QuestionRepository,
	validationService ValidationService,
) QuizService {
	quizServiceOnce.Do(func() {
		quizServiceInstance = &quizService{
			logger:            logger,
			quizRepo:          quizRepo,
			questionRepo:      questionRepo,
			validationService: validationService,
		}
	})
	return quizServiceInstance
}

func (s *quizService) CreateQuiz(ctx context.Context, authUser *dtos.UserSession, req *dtos.CreateQuizRequest) (*models.Quiz, *exception.AppError) {
	s.logger.Info("[CREATE QUIZ]", authUser, req)

	quiz := &models.Quiz{
		Title:           req.Title,
		Description:     req.Description,
		Visibility:      models.QuizVisibilityPrivate,
		OwnerID:         authUser.UserID,
		MaxParticipants: &req.MaxParticipants,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	createdQuiz, err := s.quizRepo.CreateQuiz(ctx, quiz)
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	return createdQuiz, nil
}

func (s *quizService) UpdateQuiz(ctx context.Context, authUser *dtos.UserSession, req *dtos.UpdateQuizRequest) (*models.Quiz, *exception.AppError) {
	s.logger.Info("[UPDATE QUIZ]", authUser, req)

	quiz, err := s.validationService.ValidateQuizOwnership(ctx, req.QuizID, authUser.UserID, false)
	if err != nil {
		return nil, err
	}

	quiz.Title = req.Title
	quiz.Description = req.Description

	if req.Visibility == models.QuizVisibilityPublished && quiz.Visibility != models.QuizVisibilityPublished {
		now := time.Now()
		quiz.PublishedAt = &now
	} else if req.Visibility != models.QuizVisibilityPublished && quiz.Visibility == models.QuizVisibilityPublished {
		quiz.PublishedAt = nil
	}

	quiz.Visibility = req.Visibility
	quiz.MaxParticipants = &req.MaxParticipants

	updatedQuiz, appErr := s.quizRepo.UpdateQuiz(ctx, quiz)
	if appErr != nil {
		if appErr == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrQuizNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, appErr.Error())
	}

	return updatedQuiz, nil
}

func (s *quizService) GetMyQuizDetail(ctx context.Context, authUser *dtos.UserSession, req *dtos.GetMyQuizDetailRequest) (*models.Quiz, *exception.AppError) {
	s.logger.Info("[GET MY QUIZ DETAIL]", req)

	// Validate ownership and get quiz with questions
	quiz, appErr := s.validationService.ValidateQuizOwnership(ctx, req.QuizID, authUser.UserID, true)
	if appErr != nil {
		return nil, appErr
	}

	return quiz, nil
}

func (s *quizService) GetMyQuizList(ctx context.Context, authUser *dtos.UserSession, req *dtos.GetMyQuizListRequest) (*models.MyQuizListResponse, *exception.AppError) {
	s.logger.Info("[GET MY QUIZ LIST]", authUser, req)

	offset, limit := utils.CalculateOffset(req.Page, req.Limit)

	queries := []func(context.Context) (any, error){
		func(ctx context.Context) (any, error) {
			return s.quizRepo.GetQuizListByOwner(
				ctx,
				authUser.UserID,
				req.Query,
				req.Visibility,
				limit,
				offset,
			)
		},
		func(ctx context.Context) (any, error) {
			return s.quizRepo.CountQuizListByOwner(
				ctx,
				authUser.UserID,
				req.Query,
				req.Visibility,
			)
		},
	}

	results, err := utils.RunQueriesParallel(ctx, queries)
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	quizzes := results[0].([]*models.Quiz)
	totalItems := results[1].(int64)

	return &models.MyQuizListResponse{
		Quizzes: quizzes,
		Pagination: response.OffsetPagination{
			Page:       req.Page,
			Limit:      req.Limit,
			TotalItems: totalItems,
		},
	}, nil
}

func (s *quizService) DeleteQuiz(ctx context.Context, authUser *dtos.UserSession, req *dtos.DeleteQuizRequest) *exception.AppError {
	s.logger.Info("[DELETE QUIZ]", authUser, req)

	quiz, appErr := s.validationService.ValidateQuizOwnership(ctx, req.QuizID, authUser.UserID, false)
	if appErr != nil {
		return appErr
	}

	err := s.quizRepo.DeleteQuiz(ctx, quiz.ID)
	if err != nil {
		return exception.InternalError(errors.CodeDBError, err.Error())
	}

	return nil
}

func (s *quizService) GetQuizList(ctx context.Context, req *dtos.GetQuizListRequest) (*dtos.GetQuizListResponse, *exception.AppError) {
	s.logger.Info("[GET PUBLIC QUIZ LIST]", req)

	limit := utils.CalculateLimit(req.Limit)
	sortBy := helpers.ValidateQuizSortBy(req.SortBy)

	quizzes, err := s.quizRepo.GetPublicQuizList(ctx, req.Query, req.Cursor, limit+1, &sortBy)
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	hasMore := len(quizzes) > int(limit)
	if hasMore {
		quizzes = quizzes[:limit] // Remove the extra item
	}

	var nextCursor *string
	if hasMore && len(quizzes) > 0 {
		lastQuiz := quizzes[len(quizzes)-1]

		cursorVal, cursorErr := helpers.EncodeQuizCursor(lastQuiz, sortBy)
		if cursorErr != nil {
			s.logger.Error("Failed to encode cursor", map[string]interface{}{
				"quiz_id": lastQuiz.ID,
				"sort_by": sortBy,
				"error":   cursorErr,
			})
			// Fallback
			fallbackCursor := strconv.FormatInt(lastQuiz.ID, 10)
			nextCursor = &fallbackCursor
		} else {
			nextCursor = &cursorVal
		}
	}

	return &dtos.GetQuizListResponse{
		Quizzes:    quizzes,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *quizService) GetQuizDetail(ctx context.Context, authUser *dtos.UserSession, req *dtos.GetQuizDetailRequest) (*models.Quiz, *exception.AppError) {
	s.logger.Info("[GET QUIZ DETAIL]", authUser, req)

	quiz, err := s.quizRepo.GetQuizDetail(ctx, req.QuizID, true)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrQuizNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	if quiz.Visibility != models.QuizVisibilityPublished {
		// If quiz is private/unlisted, check if user is the owner
		if authUser == nil || quiz.OwnerID != authUser.UserID {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrQuizNotFound)
		}
	}

	// Increment view count if user is not the owner
	if authUser == nil || quiz.OwnerID != authUser.UserID {
		go func() {
			if err := s.quizRepo.IncrementViewCount(context.Background(), req.QuizID); err != nil {
				s.logger.Error("Failed to increment view count", map[string]interface{}{
					"quiz_id": req.QuizID,
					"error":   err,
				})
			}
		}()
	}

	return quiz, nil
}
