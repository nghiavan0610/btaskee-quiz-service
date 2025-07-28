package errors

const (
	ErrSessionNotFound     = "Session not found"
	ErrSessionNotStartable = "Session cannot be started"
	ErrSessionNotEndable   = "Session cannot be ended"
	ErrParticipantNotFound = "Participant not found"
	ErrSessionFull         = "Session is full"
	ErrSessionAlreadyEnded = "Session has already ended"
	ErrDuplicateAnswer     = "Answer already submitted for this question"
	ErrInvalidJoinCode     = "Invalid join code"
	ErrSessionNotActive    = "Session is not active"
	ErrSessionNotJoinable  = "Session is not joinable"
)
