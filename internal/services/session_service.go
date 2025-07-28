package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/dtos"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/errors"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
)

type (
	SessionService interface {
		CreateSession(ctx context.Context, authUser *dtos.UserSession, req *dtos.CreateSessionRequest) (*dtos.CreateSessionResponse, *exception.AppError)
		JoinSession(ctx context.Context, authUser *dtos.UserSession, req *dtos.JoinSessionRequest) (*models.SessionParticipant, *exception.AppError)
	}

	sessionService struct {
		sessionRepo  repositories.SessionRepository
		quizRepo     repositories.QuizRepository
		questionRepo repositories.QuestionRepository
		logger       *logger.Logger
	}
)

var (
	sessionServiceOnce     sync.Once
	sessionServiceInstance SessionService
)

func ProvideSessionService(
	sessionRepo repositories.SessionRepository,
	quizRepo repositories.QuizRepository,
	questionRepo repositories.QuestionRepository,
	logger *logger.Logger,
) SessionService {
	sessionServiceOnce.Do(func() {
		sessionServiceInstance = &sessionService{
			sessionRepo:  sessionRepo,
			quizRepo:     quizRepo,
			questionRepo: questionRepo,
			logger:       logger,
		}
	})
	return sessionServiceInstance
}

func (s *sessionService) CreateSession(ctx context.Context, authUser *dtos.UserSession, req *dtos.CreateSessionRequest) (*dtos.CreateSessionResponse, *exception.AppError) {
	quiz, err := s.quizRepo.GetQuizDetail(ctx, req.QuizID, false)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrQuizNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	// Check if quiz is published (public quizzes can be hosted by anyone)
	if quiz.Visibility != models.QuizVisibilityPublished && (authUser == nil || quiz.OwnerID != authUser.UserID) {
		return nil, exception.Forbidden(errors.CodeForbidden, errors.ErrForbidden).WithDetails("You don't have permission to host this quiz")
	}

	// Generate unique join code
	joinCode, err := s.sessionRepo.GenerateJoinCode(ctx)
	if err != nil {
		return nil, exception.InternalError(errors.CodeInternal, err.Error())
	}

	// Create session
	var hostID *int64
	var hostName string = "Anonymous Host"
	if authUser != nil {
		hostID = &authUser.UserID
		hostName = authUser.Username
	} else {
		anonymousID := int64(1000000000) + time.Now().Unix()
		hostID = &anonymousID
	}

	session := &models.QuizSession{
		QuizID:               req.QuizID,
		HostID:               hostID,
		JoinCode:             joinCode,
		Status:               models.SessionStatusWaiting,
		MaxParticipants:      quiz.MaxParticipants,
		CurrentQuestionIndex: 0,
		ParticipantCount:     0,
	}

	createdSession, err := s.sessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, exception.InternalError(errors.CodeInternal, err.Error())
	}

	hostParticipant := &models.SessionParticipant{
		SessionID: createdSession.ID,
		UserID:    hostID,
		Nickname:  hostName,
		Score:     0,
		IsHost:    true,
	}

	_, err = s.sessionRepo.AddParticipant(ctx, hostParticipant)
	if err != nil {
		s.logger.Error("Failed to add host as participant", err)
	} else {
		createdSession.ParticipantCount = 1
		_, updateErr := s.sessionRepo.UpdateSession(ctx, createdSession)
		if updateErr != nil {
			s.logger.Error("Failed to update participant count after adding host", updateErr)
		}
	}

	response := &dtos.CreateSessionResponse{
		SessionID:       createdSession.ID,
		JoinCode:        createdSession.JoinCode,
		JoinURL:         fmt.Sprintf("/games/join/%s", createdSession.JoinCode),
		QuizTitle:       quiz.Title,
		HostName:        hostName,
		HostUserID:      hostID,
		IsHost:          true,
		MaxParticipants: quiz.MaxParticipants,
	}

	return response, nil
}

func (s *sessionService) JoinSession(ctx context.Context, authUser *dtos.UserSession, req *dtos.JoinSessionRequest) (*models.SessionParticipant, *exception.AppError) {
	session, err := s.sessionRepo.GetSessionByJoinCode(ctx, req.JoinCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, exception.NotFound(errors.CodeNotFound, errors.ErrSessionNotFound)
		}
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	// Check if session is joinable
	if session.Status != models.SessionStatusWaiting {
		return nil, exception.BadRequest(errors.CodeBadRequest, errors.ErrSessionNotJoinable)
	}

	// Check participant limit
	if session.MaxParticipants != nil && session.ParticipantCount >= *session.MaxParticipants {
		return nil, exception.BadRequest(errors.CodeBadRequest, errors.ErrSessionFull)
	}

	// Check if user is already a participant (including if they're the host)
	if authUser != nil {
		participants, err := s.sessionRepo.GetSessionParticipants(ctx, session.ID)
		if err == nil {
			for _, participant := range participants {
				if participant.UserID != nil && *participant.UserID == authUser.UserID {
					// User is already in the session, return existing participant
					return participant, nil
				}
			}
		}
	}

	nickname := "Btaskee's Guest"
	if authUser != nil {
		nickname = authUser.Username
	}

	// Create participant
	var userID *int64
	if authUser != nil {
		userID = &authUser.UserID
	} else {
		// Generate unique anonymous ID using high positive range (1 billion + timestamp in seconds)
		// This avoids JavaScript precision issues and conflicts with real user IDs
		anonymousID := int64(1000000000) + time.Now().Unix()
		userID = &anonymousID
	}

	participant := &models.SessionParticipant{
		SessionID: session.ID,
		UserID:    userID,
		Nickname:  nickname,
		Score:     0,
		IsHost:    false,
	}

	createdParticipant, err := s.sessionRepo.AddParticipant(ctx, participant)
	if err != nil {
		return nil, exception.InternalError(errors.CodeDBError, err.Error())
	}

	session.ParticipantCount += 1
	_, err = s.sessionRepo.UpdateSession(ctx, session)
	if err != nil {
		s.logger.Error("Failed to update participant count", err)
	}

	return createdParticipant, nil
}
