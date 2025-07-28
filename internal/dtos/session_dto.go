package dtos

type CreateSessionRequest struct {
	QuizID int64 `json:"quiz_id" validate:"required,min=1"`
}

type CreateSessionResponse struct {
	SessionID       int64  `json:"session_id"`
	JoinCode        string `json:"join_code"`
	JoinURL         string `json:"join_url"`
	QuizTitle       string `json:"quiz_title"`
	HostName        string `json:"host_name"`
	HostUserID      *int64 `json:"host_user_id"`
	IsHost          bool   `json:"is_host"`
	MaxParticipants *int32 `json:"max_participants,omitempty"`
}

type JoinSessionRequest struct {
	JoinCode string `params:"join_code" validate:"required,len=6"`
}
