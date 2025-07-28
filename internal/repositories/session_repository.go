package repositories

import (
	"context"
	"fmt"

	helpers "github.com/nghiavan0610/btaskee-quiz-service/helpers/session"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/transformers"
)

type SessionRepository interface {
	// HTTP API (Initial Setup Only)
	CreateSession(ctx context.Context, session *models.QuizSession) (*models.QuizSession, error)
	GetSessionByID(ctx context.Context, sessionID int64) (*models.QuizSession, error)
	GetSessionByJoinCode(ctx context.Context, joinCode string) (*models.QuizSession, error)
	AddParticipant(ctx context.Context, participant *models.SessionParticipant) (*models.SessionParticipant, error)
	GenerateJoinCode(ctx context.Context) (string, error)
	UpdateSession(ctx context.Context, session *models.QuizSession) (*models.QuizSession, error)

	// WebSocket API (Real-time Gameplay)
	StartSession(ctx context.Context, sessionID int64) error
	EndSession(ctx context.Context, sessionID int64) error
	UpdateSessionQuestion(ctx context.Context, sessionID int64, questionIndex int32) error
	GetSessionParticipants(ctx context.Context, sessionID int64) ([]*models.SessionParticipant, error)
	GetSessionLeaderboard(ctx context.Context, sessionID int64) ([]*models.LeaderboardParticipant, error)
	UpdateParticipantScore(ctx context.Context, participantID int64, score int32) error
}

type sessionRepository struct {
	queries *sqlc.Queries
}

func ProvideSessionRepository(queries *sqlc.Queries) SessionRepository {
	return &sessionRepository{
		queries: queries,
	}
}

// ===== HTTP API (Initial Setup Only) =====

func (r *sessionRepository) GetSessionByID(ctx context.Context, sessionID int64) (*models.QuizSession, error) {
	result, err := r.queries.GetSessionByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	session := &models.QuizSession{
		ID:                   result.ID,
		QuizID:               result.QuizID,
		HostID:               result.HostID,
		JoinCode:             result.JoinCode,
		Status:               models.SessionStatus(result.Status),
		CurrentQuestionIndex: result.CurrentQuestionIndex,
		MaxParticipants:      result.MaxParticipants,
		ParticipantCount:     result.ParticipantCount,
		StartedAt:            result.StartedAt.Time,
		EndedAt:              result.EndedAt.Time,
		CreatedAt:            result.CreatedAt.Time,
		UpdatedAt:            result.UpdatedAt.Time,
		Quiz: &models.Quiz{
			ID:             result.QuizID,
			Title:          result.QuizTitle,
			Description:    result.QuizDescription,
			TotalQuestions: result.QuizTotalQuestions,
		},
	}

	return session, nil
}

func (r *sessionRepository) CreateSession(ctx context.Context, session *models.QuizSession) (*models.QuizSession, error) {
	params := sqlc.CreateSessionParams{
		QuizID:               session.QuizID,
		HostID:               session.HostID,
		JoinCode:             session.JoinCode,
		Status:               sqlc.SessionStatus(session.Status),
		MaxParticipants:      session.MaxParticipants,
		CurrentQuestionIndex: session.CurrentQuestionIndex,
		ParticipantCount:     session.ParticipantCount,
	}

	result, err := r.queries.CreateSession(ctx, params)
	if err != nil {
		return nil, err
	}

	return transformers.ConvertSQLCSessionToModel(result), nil
}

func (r *sessionRepository) GetSessionByJoinCode(ctx context.Context, joinCode string) (*models.QuizSession, error) {
	result, err := r.queries.GetSessionByJoinCode(ctx, joinCode)
	if err != nil {
		return nil, err
	}

	session := &models.QuizSession{
		ID:                   result.ID,
		QuizID:               result.QuizID,
		HostID:               result.HostID,
		JoinCode:             result.JoinCode,
		Status:               models.SessionStatus(result.Status),
		CurrentQuestionIndex: result.CurrentQuestionIndex,
		MaxParticipants:      result.MaxParticipants,
		ParticipantCount:     result.ParticipantCount,
		StartedAt:            result.StartedAt.Time,
		EndedAt:              result.EndedAt.Time,
		CreatedAt:            result.CreatedAt.Time,
		UpdatedAt:            result.UpdatedAt.Time,
		Quiz: &models.Quiz{
			ID:             result.QuizID,
			Title:          result.QuizTitle,
			Description:    result.QuizDescription,
			TotalQuestions: result.QuizTotalQuestions,
		},
	}

	return session, nil
}

func (r *sessionRepository) AddParticipant(ctx context.Context, participant *models.SessionParticipant) (*models.SessionParticipant, error) {
	params := sqlc.AddParticipantParams{
		SessionID: participant.SessionID,
		UserID:    participant.UserID,
		Nickname:  participant.Nickname,
		Score:     participant.Score,
		IsHost:    participant.IsHost,
	}

	result, err := r.queries.AddParticipant(ctx, params)
	if err != nil {
		return nil, err
	}

	return transformers.ConvertSQLCParticipantToModel(result), nil
}

func (r *sessionRepository) GenerateJoinCode(ctx context.Context) (string, error) {
	maxRetries := 10

	for range maxRetries {
		code, err := helpers.GenerateJoinCode()
		if err != nil {
			return "", err
		}

		exists, err := r.queries.CheckJoinCodeExists(ctx, code)
		if err != nil {
			return "", err
		}

		if !exists {
			return code, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique join code after %d attempts", maxRetries)
}

func (r *sessionRepository) UpdateSession(ctx context.Context, session *models.QuizSession) (*models.QuizSession, error) {
	params := sqlc.UpdateSessionParams{
		ID:                   session.ID,
		Status:               sqlc.SessionStatus(session.Status),
		CurrentQuestionIndex: session.CurrentQuestionIndex,
		MaxParticipants:      session.MaxParticipants,
		ParticipantCount:     session.ParticipantCount,
	}

	result, err := r.queries.UpdateSession(ctx, params)
	if err != nil {
		return nil, err
	}

	return transformers.ConvertSQLCSessionToModel(result), nil
}

// ===== WebSocket API (Real-time Gameplay) =====

func (r *sessionRepository) StartSession(ctx context.Context, sessionID int64) error {
	return r.queries.StartSession(ctx, sessionID)
}

func (r *sessionRepository) EndSession(ctx context.Context, sessionID int64) error {
	return r.queries.EndSession(ctx, sessionID)
}

func (r *sessionRepository) UpdateSessionQuestion(ctx context.Context, sessionID int64, questionIndex int32) error {
	params := sqlc.UpdateSessionQuestionParams{
		ID:                   sessionID,
		CurrentQuestionIndex: questionIndex,
	}
	return r.queries.UpdateSessionQuestion(ctx, params)
}

func (r *sessionRepository) GetSessionParticipants(ctx context.Context, sessionID int64) ([]*models.SessionParticipant, error) {
	result, err := r.queries.GetSessionParticipants(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	participants := make([]*models.SessionParticipant, len(result))
	for i, p := range result {
		participants[i] = transformers.ConvertSQLCParticipantToModel(p)
	}

	return participants, nil
}

func (r *sessionRepository) GetSessionLeaderboard(ctx context.Context, sessionID int64) ([]*models.LeaderboardParticipant, error) {
	result, err := r.queries.GetSessionLeaderboard(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	leaderboard := make([]*models.LeaderboardParticipant, len(result))
	for i, row := range result {
		leaderboard[i] = transformers.ConvertSQLCSessionLeaderboardToModel(row)
	}

	return leaderboard, nil
}

func (r *sessionRepository) UpdateParticipantScore(ctx context.Context, participantID int64, score int32) error {
	params := sqlc.UpdateParticipantScoreParams{
		ID:    participantID,
		Score: score,
	}
	return r.queries.UpdateParticipantScore(ctx, params)
}
