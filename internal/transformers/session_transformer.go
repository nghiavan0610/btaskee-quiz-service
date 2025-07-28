package transformers

import (
	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
)

func ConvertSQLCSessionToModel(session sqlc.QuizSession) *models.QuizSession {
	return &models.QuizSession{
		ID:                   session.ID,
		QuizID:               session.QuizID,
		HostID:               session.HostID,
		JoinCode:             session.JoinCode,
		Status:               models.SessionStatus(session.Status),
		CurrentQuestionIndex: session.CurrentQuestionIndex,
		MaxParticipants:      session.MaxParticipants,
		ParticipantCount:     session.ParticipantCount,
		StartedAt:            session.StartedAt.Time,
		EndedAt:              session.EndedAt.Time,
		CreatedAt:            session.CreatedAt.Time,
		UpdatedAt:            session.UpdatedAt.Time,
	}
}

func ConvertSQLCParticipantToModel(participant sqlc.SessionParticipant) *models.SessionParticipant {
	return &models.SessionParticipant{
		ID:           participant.ID,
		SessionID:    participant.SessionID,
		UserID:       participant.UserID,
		Nickname:     participant.Nickname,
		Score:        participant.Score,
		IsHost:       participant.IsHost,
		JoinedAt:     participant.JoinedAt.Time,
		LastActivity: participant.LastActivity.Time,
	}
}

func ConvertSQLCSessionLeaderboardToModel(row sqlc.GetSessionLeaderboardRow) *models.LeaderboardParticipant {
	return &models.LeaderboardParticipant{
		ID:       row.ID,
		Nickname: row.Nickname,
		Score:    row.Score,
		IsHost:   row.IsHost,
		Rank:     row.Rank,
	}
}
