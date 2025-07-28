package models

import (
	"time"
)

type SessionStatus string

const (
	SessionStatusWaiting   SessionStatus = "waiting"
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusCancelled SessionStatus = "cancelled"
)

type QuizSession struct {
	ID                   int64         `json:"id"`
	QuizID               int64         `json:"quiz_id"`
	HostID               *int64        `json:"host_id,omitempty"`
	JoinCode             string        `json:"join_code"`
	Status               SessionStatus `json:"status"`
	CurrentQuestionIndex int32         `json:"current_question_index"`
	MaxParticipants      *int32        `json:"max_participants,omitempty"`
	ParticipantCount     int32         `json:"participant_count"`
	StartedAt            time.Time     `json:"started_at,omitempty"`
	EndedAt              time.Time     `json:"ended_at,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`

	// Related data
	Quiz         *Quiz                 `json:"quiz,omitempty"`
	Host         *User                 `json:"host,omitempty"`
	Participants []*SessionParticipant `json:"participants,omitempty"`
	Leaderboard  []*LeaderboardEntry   `json:"leaderboard,omitempty"`
}

type SessionParticipant struct {
	ID           int64     `json:"id"`
	SessionID    int64     `json:"session_id"`
	UserID       *int64    `json:"user_id,omitempty"`
	Nickname     string    `json:"nickname"`
	Score        int32     `json:"score"`
	IsHost       bool      `json:"is_host"`
	JoinedAt     time.Time `json:"joined_at"`
	LastActivity time.Time `json:"last_activity"`
	User         *User     `json:"user,omitempty"`
}

type LeaderboardEntry struct {
	ID             int64               `json:"id"`
	SessionID      int64               `json:"session_id"`
	ParticipantID  int64               `json:"participant_id"`
	Rank           int32               `json:"rank"`
	Score          int32               `json:"score"`
	TotalAnswers   int32               `json:"total_answers"`
	CorrectAnswers int32               `json:"correct_answers"`
	AverageTime    float64             `json:"average_time"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Participant    *SessionParticipant `json:"participant,omitempty"`
}

type LeaderboardParticipant struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Score    int32  `json:"score"`
	IsHost   bool   `json:"is_host"`
	Rank     int64  `json:"rank"`
}
