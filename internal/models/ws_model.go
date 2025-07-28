package models

import "time"

type WSMessageType string

const (
	// Client to Server
	WSMsgTypeJoin         WSMessageType = "join"
	WSMsgTypeAnswer       WSMessageType = "answer"
	WSMsgTypeStartQuiz    WSMessageType = "start_quiz"
	WSMsgTypeNextQuestion WSMessageType = "next_question"
	WSMsgTypeEndQuiz      WSMessageType = "end_quiz"
	WSMsgTypePauseQuiz    WSMessageType = "pause_quiz"
	WSMsgTypeResumeQuiz   WSMessageType = "resume_quiz"
	WSMsgTypePing         WSMessageType = "ping"
	WSMsgTypeGetState     WSMessageType = "get_session_state"

	// Server to Client
	WSMsgTypeJoinSuccess           WSMessageType = "join_success"
	WSMsgTypeParticipantJoin       WSMessageType = "participant_join"
	WSMsgTypeParticipantLeft       WSMessageType = "participant_left"
	WSMsgTypeParticipantListUpdate WSMessageType = "participant_list_update"
	WSMsgTypeQuizStart             WSMessageType = "quiz_start"
	WSMsgTypeGameStarted           WSMessageType = "game_started"
	WSMsgTypeGamePaused            WSMessageType = "game_paused"
	WSMsgTypeGameResumed           WSMessageType = "game_resumed"
	WSMsgTypeGameEnded             WSMessageType = "game_ended"
	WSMsgTypeQuestionStart         WSMessageType = "question_start"
	WSMsgTypeQuestionEnd           WSMessageType = "question_end"
	WSMsgTypeAnswerReceived        WSMessageType = "answer_received"
	WSMsgTypeScoreUpdate           WSMessageType = "score_update"
	WSMsgTypeLeaderboard           WSMessageType = "leaderboard"
	WSMsgTypeQuizEnd               WSMessageType = "quiz_end"
	WSMsgTypeError                 WSMessageType = "error"
	WSMsgTypePong                  WSMessageType = "pong"
	WSMsgTypeSessionState          WSMessageType = "session_state"
)

type WSMessage struct {
	Type      WSMessageType `json:"type"`
	Payload   interface{}   `json:"payload"`
	Timestamp time.Time     `json:"timestamp"`
}

type WSJoinPayload struct {
	SessionID string `json:"session_id,omitempty"`
	Nickname  string `json:"nickname"`
	UserID    *int64 `json:"user_id,omitempty"`
}

type WSJoinSuccessPayload struct {
	Session     *QuizSession        `json:"session"`
	Participant *SessionParticipant `json:"participant"`
	IsHost      bool                `json:"is_host"`
}

type WSAnswerPayload struct {
	QuestionID   int64  `json:"question_id"`
	QuestionType string `json:"question_type"` // "single_choice", "multiple_choice", "text_input"
	// For single choice and text input
	AnswerValue *string `json:"answer_value,omitempty"`
	// For multiple choice
	AnswerValues []string `json:"answer_values,omitempty"`
	TimeTaken    int32    `json:"time_taken"` // milliseconds
}

type WSAnswerReceivedPayload struct {
	IsCorrect   bool  `json:"is_correct"`
	ScoreEarned int32 `json:"score_earned"`
	TimeTaken   int32 `json:"time_taken"`
}
