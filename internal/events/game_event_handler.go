package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/repositories"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/constants"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/exception"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
	ws "github.com/nghiavan0610/btaskee-quiz-service/pkg/websocket"
	"github.com/nghiavan0610/btaskee-quiz-service/utils"
)

type GameEventHandler interface {
	// WebSocket message handler implementation
	ws.WebSocketMessageHandler

	// Game lifecycle events
	NotifyGameStart(sessionID int64) error
	NotifyQuestionStart(sessionID int64, question *models.Question) error
	NotifyQuestionEnd(sessionID int64) error
	NotifyParticipantJoin(sessionID int64, participant *models.SessionParticipant) error
	NotifyParticipantLeft(sessionID int64, participantID int64) error

	// Client management
	HandleClientDisconnect(client *ws.Client)
}

type gameEventHandler struct {
	sessionRepo    repositories.SessionRepository
	questionRepo   repositories.QuestionRepository
	hub            ws.Hub
	logger         *logger.Logger
	questionTimers map[int64]*time.Timer // sessionID -> timer for auto progression
	timerMutex     sync.RWMutex          // protect timer map

	// PERFORMANCE: Message marshaling cache for frequently sent messages
	marshalCache sync.Map // string -> []byte (cached marshaled messages)
}

var (
	gameEventHandlerOnce     sync.Once
	gameEventHandlerInstance GameEventHandler
)

func ProvideGameEventHandler(
	sessionRepo repositories.SessionRepository,
	questionRepo repositories.QuestionRepository,
	hub ws.Hub,
	logger *logger.Logger,
) GameEventHandler {
	gameEventHandlerOnce.Do(func() {
		gameEventHandlerInstance = &gameEventHandler{
			sessionRepo:    sessionRepo,
			questionRepo:   questionRepo,
			hub:            hub,
			logger:         logger,
			questionTimers: make(map[int64]*time.Timer),
		}
	})

	// Register this service as the WebSocket message handler
	hub.SetMessageHandler(gameEventHandlerInstance)

	return gameEventHandlerInstance
}

// HandleMessage implements WebSocketMessageHandler
func (s *gameEventHandler) HandleMessage(client *ws.Client, wsMsg *models.WSMessage) {
	// Fast path routing for most common messages
	switch wsMsg.Type {
	case models.WSMsgTypeAnswer:
		s.handleSubmitAnswer(client, wsMsg)
	case models.WSMsgTypePing:
		s.handlePing(client, wsMsg)
	case models.WSMsgTypeJoin:
		s.handleJoinRoom(client, wsMsg)
	case models.WSMsgTypeStartQuiz:
		s.handleStartGame(client, wsMsg)
	case models.WSMsgTypeNextQuestion:
		s.handleNextQuestion(client, wsMsg)
	case models.WSMsgTypeEndQuiz:
		s.handleEndGame(client, wsMsg)
	case models.WSMsgTypePauseQuiz:
		s.handlePauseGame(client, wsMsg)
	case models.WSMsgTypeResumeQuiz:
		s.handleResumeGame(client, wsMsg)
	case models.WSMsgTypeGetState:
		s.handleGetSessionState(client, wsMsg) // Get current session state for client synchronization
	default:
		s.logger.Warn("Unknown WebSocket message type", map[string]interface{}{
			"client_id": client.ID,
			"type":      wsMsg.Type,
		})
	}
}

func (s *gameEventHandler) handleJoinRoom(client *ws.Client, wsMsg *models.WSMessage) {
	ctx := context.Background()

	// Parse typed payload
	var payload models.WSJoinPayload
	if err := s.parsePayload(wsMsg.Payload, &payload); err != nil {
		s.sendError(client, "INVALID_PAYLOAD", "Invalid join message payload format")
		return
	}

	// Require user_id from HTTP API (CreateSession/JoinSession)
	if payload.UserID == nil {
		s.sendError(client, "USER_ID_REQUIRED", "Please call CreateSession or JoinSession API first to get user_id")
		return
	}

	// Get session and find the participant by user_id
	queries := []func(context.Context) (any, error){
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionByID(ctx, client.SessionID)
		},
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionParticipants(ctx, client.SessionID)
		},
	}

	results, err := utils.RunQueriesParallel(ctx, queries)
	if err != nil {
		s.sendError(client, "SESSION_NOT_FOUND", "Session not found")
		return
	}

	session := results[0].(*models.QuizSession)
	participants := results[1].([]*models.SessionParticipant)

	// Find participant by user_id (created via HTTP API)
	var participant *models.SessionParticipant
	for _, p := range participants {
		if p.UserID != nil && *p.UserID == *payload.UserID {
			participant = p
			break
		}
	}

	if participant == nil {
		s.sendError(client, "PARTICIPANT_NOT_FOUND", "No participant found with this user_id. Please call CreateSession or JoinSession API first")
		return
	}

	// Set client participant info
	client.SetParticipantInfo(participant.ID, participant.UserID, participant.Nickname, participant.IsHost)

	successPayload := &models.WSJoinSuccessPayload{
		Session:     session,
		Participant: participant,
		IsHost:      participant.IsHost,
	}

	response := &models.WSMessage{
		Type:      models.WSMsgTypeJoinSuccess,
		Payload:   successPayload,
		Timestamp: time.Now(),
	}

	s.sendToClient(client, response)

	go func() {
		// Small delay to ensure join_success is processed first
		time.Sleep(100 * time.Millisecond)

		s.NotifyParticipantJoin(client.SessionID, participant)

		s.broadcastParticipantListUpdate(client.SessionID, "websocket_connection", nil)
	}()
}

func (s *gameEventHandler) handleSubmitAnswer(client *ws.Client, wsMsg *models.WSMessage) {
	var answerPayload models.WSAnswerPayload
	if err := s.parsePayload(wsMsg.Payload, &answerPayload); err != nil {
		s.sendError(client, "INVALID_PAYLOAD", "Invalid answer payload format")
		return
	}

	ctx := context.Background()

	question, err := s.questionRepo.GetQuestionByID(ctx, answerPayload.QuestionID)
	if err != nil {
		s.sendError(client, "QUESTION_NOT_FOUND", "Question not found")
		return
	}

	// Validate payload based on question type
	switch question.Type {
	case models.QuestionTypeSingleChoice, models.QuestionTypeTextInput:
		if answerPayload.AnswerValue == nil || *answerPayload.AnswerValue == "" {
			s.sendError(client, "INVALID_ANSWER", "Answer value cannot be empty for single choice/text input")
			return
		}
	case models.QuestionTypeMultipleChoice:
		if len(answerPayload.AnswerValues) == 0 {
			s.sendError(client, "INVALID_ANSWER", "Answer values cannot be empty for multiple choice")
			return
		}
	}

	// Calculate score
	isCorrect := s.evaluateAnswer(question, &answerPayload)

	scoreEarned := int32(0)
	if isCorrect {
		// Score based on time: max 1000 points, reduced by time taken
		baseScore := int32(constants.MaxQuestionScore)
		timePenalty := int32(answerPayload.TimeTaken / 100) // 1 point per 100ms
		scoreEarned = baseScore - timePenalty
		if scoreEarned < constants.MinQuestionScore {
			scoreEarned = constants.MinQuestionScore // minimum score for correct answer
		}
	}

	// Update score and send response
	err = s.sessionRepo.UpdateParticipantScore(ctx, client.ParticipantID, scoreEarned)
	if err != nil {
		s.sendError(client, "SCORE_UPDATE_FAILED", "Failed to update score")
		return
	}

	// Send answer received confirmation
	responsePayload := &models.WSAnswerReceivedPayload{
		IsCorrect:   isCorrect,
		ScoreEarned: scoreEarned,
		TimeTaken:   answerPayload.TimeTaken,
	}

	response := &models.WSMessage{
		Type:      models.WSMsgTypeAnswerReceived,
		Payload:   responsePayload,
		Timestamp: time.Now(),
	}
	s.sendToClient(client, response)

	// Add async leaderboard broadcast if needed
	// go s.broadcastLeaderboard(context.Background(), client.SessionID)
}

func (s *gameEventHandler) handleStartGame(client *ws.Client, wsMsg *models.WSMessage) {
	if !client.IsHost {
		s.sendError(client, "UNAUTHORIZED", "Only host can start the game")
		return
	}

	ctx := context.Background()

	queries := []func(context.Context) (any, error){
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionByID(ctx, client.SessionID)
		},
		func(ctx context.Context) (any, error) {
			session, err := s.sessionRepo.GetSessionByID(ctx, client.SessionID)
			if err != nil {
				return nil, err
			}
			return s.questionRepo.GetQuestionListByQuiz(ctx, session.QuizID)
		},
	}

	results, err := utils.RunQueriesParallel(ctx, queries)
	if err != nil {
		s.sendError(client, "SESSION_NOT_FOUND", "Session not found")
		return
	}

	session := results[0].(*models.QuizSession)
	questions := results[1].([]*models.Question)

	if session.Status != models.SessionStatusWaiting {
		s.sendError(client, "INVALID_STATUS", "Session cannot be started from current status")
		return
	}

	if len(questions) == 0 {
		s.sendError(client, "NO_QUESTIONS", "No questions available for this quiz")
		return
	}

	// 2. Update session status to started
	err = s.sessionRepo.StartSession(ctx, client.SessionID)
	if err != nil {
		s.logger.Error("Failed to start session", err)
		s.sendError(client, "START_FAILED", "Failed to start session")
		return
	}

	s.logger.Info("Game started successfully", map[string]interface{}{
		"session_id":      client.SessionID,
		"host_id":         client.ID,
		"questions_count": len(questions),
	})

	// 3. Broadcast game start to all participants
	s.NotifyGameStart(client.SessionID)

	// Automatically start the first question after game start
	go func() {
		// Small delay to ensure game start message is processed first
		time.Sleep(1 * time.Second)

		// Update session to first question (index 0)
		err := s.sessionRepo.UpdateSessionQuestion(ctx, client.SessionID, 0)
		if err != nil {
			s.logger.Error("Failed to update session to first question", err)
			return
		}

		// Start the first question
		firstQuestion := questions[0]

		s.NotifyQuestionStart(client.SessionID, firstQuestion)
	}()
}

func (s *gameEventHandler) handleNextQuestion(client *ws.Client, wsMsg *models.WSMessage) {
	if !client.IsHost {
		s.sendError(client, "UNAUTHORIZED", "Only host can control questions")
		return
	}

	ctx := context.Background()

	queries := []func(context.Context) (any, error){
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionByID(ctx, client.SessionID)
		},
		func(ctx context.Context) (any, error) {
			session, err := s.sessionRepo.GetSessionByID(ctx, client.SessionID)
			if err != nil {
				return nil, err
			}
			return s.questionRepo.GetQuestionListByQuiz(ctx, session.QuizID)
		},
	}

	results, err := utils.RunQueriesParallel(ctx, queries)
	if err != nil {
		s.sendError(client, "SESSION_NOT_FOUND", "Session not found")
		return
	}

	session := results[0].(*models.QuizSession)
	questions := results[1].([]*models.Question)

	if session.Status != models.SessionStatusActive {
		s.sendError(client, "INVALID_STATUS", "Session is not active")
		return
	}

	// 2. End current question if needed
	if session.CurrentQuestionIndex >= 0 {
		s.NotifyQuestionEnd(client.SessionID)
	}

	// 3. Get next question index
	nextQuestionIndex := session.CurrentQuestionIndex + 1

	// 5. Check if we've reached the end (questions already loaded)
	if nextQuestionIndex >= int32(len(questions)) {
		s.sendError(client, "NO_MORE_QUESTIONS", "No more questions available")
		return
	}

	// 6. Update session with new question index
	err = s.sessionRepo.UpdateSessionQuestion(ctx, client.SessionID, nextQuestionIndex)
	if err != nil {
		s.logger.Error("Failed to update session question index", err)
		s.sendError(client, "UPDATE_FAILED", "Failed to update question")
		return
	}

	// 7. Get the new question
	currentQuestion := questions[nextQuestionIndex]

	s.logger.Info("Next question started", map[string]interface{}{
		"session_id":     client.SessionID,
		"question_index": nextQuestionIndex,
		"question_id":    currentQuestion.ID,
	})

	// 8. Broadcast new question to all participants
	s.NotifyQuestionStart(client.SessionID, currentQuestion)
}

func (s *gameEventHandler) handleEndGame(client *ws.Client, wsMsg *models.WSMessage) {
	if !client.IsHost {
		s.sendError(client, "UNAUTHORIZED", "Only host can end the game")
		return
	}

	ctx := context.Background()

	// 1. Get current session
	session, err := s.sessionRepo.GetSessionByID(ctx, client.SessionID)
	if err != nil {
		s.sendError(client, "SESSION_NOT_FOUND", "Session not found")
		return
	}

	if session.Status == models.SessionStatusCompleted {
		s.sendError(client, "ALREADY_ENDED", "Game has already ended")
		return
	}

	// Stop any running timers
	s.stopQuestionTimer(client.SessionID)

	// 2. End current question if active
	if session.Status == models.SessionStatusActive {
		s.NotifyQuestionEnd(client.SessionID)
	}

	// 3. End session and get final leaderboard in parallel
	queries := []func(context.Context) (any, error){
		func(ctx context.Context) (any, error) {
			return nil, s.sessionRepo.EndSession(ctx, client.SessionID)
		},
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionLeaderboard(ctx, client.SessionID)
		},
	}

	results, err := utils.RunQueriesParallel(ctx, queries)
	if err != nil {
		s.logger.Error("Failed to end session", err)
		s.sendError(client, "END_FAILED", "Failed to end session")
		return
	}

	// 4. Get final leaderboard (from parallel execution)
	finalLeaderboard := results[1].([]*models.LeaderboardParticipant)

	s.logger.Info("Game ended successfully", map[string]interface{}{
		"session_id":         client.SessionID,
		"host_id":            client.ID,
		"participants_count": len(finalLeaderboard),
	})

	// 5. Broadcast game end with final results
	message := &models.WSMessage{
		Type: models.WSMsgTypeQuizEnd,
		Payload: map[string]interface{}{
			"session_id":        client.SessionID,
			"ended_at":          time.Now(),
			"final_leaderboard": s.formatLeaderboard(finalLeaderboard),
			"status":            "completed",
		},
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(client.SessionID, message, nil)
}

func (s *gameEventHandler) handlePing(client *ws.Client, wsMsg *models.WSMessage) {
	response := &models.WSMessage{
		Type:      models.WSMsgTypePong,
		Payload:   nil,
		Timestamp: time.Now(),
	}

	s.sendToClient(client, response)
}

func (s *gameEventHandler) handlePauseGame(client *ws.Client, wsMsg *models.WSMessage) {
	if !client.IsHost {
		s.sendError(client, "UNAUTHORIZED", "Only host can pause the game")
		return
	}

	message := &models.WSMessage{
		Type: models.WSMsgTypeGamePaused,
		Payload: map[string]interface{}{
			"session_id": client.SessionID,
			"paused_at":  time.Now(),
		},
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(client.SessionID, message, nil)
}

func (s *gameEventHandler) handleResumeGame(client *ws.Client, wsMsg *models.WSMessage) {
	if !client.IsHost {
		s.sendError(client, "UNAUTHORIZED", "Only host can resume the game")
		return
	}

	message := &models.WSMessage{
		Type: models.WSMsgTypeGameResumed,
		Payload: map[string]interface{}{
			"session_id": client.SessionID,
			"resumed_at": time.Now(),
		},
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(client.SessionID, message, nil)
}

func (s *gameEventHandler) handleGetSessionState(client *ws.Client, wsMsg *models.WSMessage) {
	ctx := context.Background()

	queries := []func(context.Context) (any, error){
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionByID(ctx, client.SessionID)
		},
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionParticipants(ctx, client.SessionID)
		},
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionLeaderboard(ctx, client.SessionID)
		},
	}

	results, err := utils.RunQueriesParallel(ctx, queries)
	if err != nil {
		s.sendError(client, "SESSION_NOT_FOUND", "Session not found")
		return
	}

	session := results[0].(*models.QuizSession)
	participants := results[1].([]*models.SessionParticipant)
	leaderboard := results[2].([]*models.LeaderboardParticipant)

	// Get current question if session is active
	var currentQuestion map[string]interface{}
	if session.Status == models.SessionStatusActive && session.CurrentQuestionIndex >= 0 {
		questions, err := s.questionRepo.GetQuestionListByQuiz(ctx, session.QuizID)
		if err == nil && int(session.CurrentQuestionIndex) < len(questions) {
			question := questions[session.CurrentQuestionIndex]
			currentQuestion = map[string]interface{}{
				"id":         question.ID,
				"question":   question.Question,
				"type":       question.Type,
				"time_limit": s.timeLimitToSeconds(question.TimeLimit),
				"index":      question.Index,
				"answers": func() []map[string]interface{} {
					answers := make([]map[string]interface{}, len(question.Answers))
					for i, answer := range question.Answers {
						answers[i] = map[string]interface{}{
							"text": answer.Text,
						}
					}
					return answers
				}(),
			}
		}
	}

	participantList := make([]map[string]interface{}, len(participants))
	for i, participant := range participants {
		participantList[i] = map[string]interface{}{
			"id":       participant.ID,
			"nickname": participant.Nickname,
			"score":    participant.Score,
			"is_host":  participant.IsHost,
		}
	}

	response := &models.WSMessage{
		Type: models.WSMsgTypeSessionState,
		Payload: map[string]interface{}{
			"session": map[string]interface{}{
				"id":                     session.ID,
				"status":                 session.Status,
				"current_question_index": session.CurrentQuestionIndex,
				"participant_count":      session.ParticipantCount,
				"max_participants":       session.MaxParticipants,
			},
			"participants":     participantList,
			"leaderboard":      s.formatLeaderboard(leaderboard),
			"current_question": currentQuestion,
		},
		Timestamp: time.Now(),
	}

	s.sendToClient(client, response)
}

func (s *gameEventHandler) HandleClientDisconnect(client *ws.Client) {
	if client.ParticipantID == 0 {
		// Client never properly joined, nothing to clean up
		return
	}

	s.logger.Info("Handling client disconnect", map[string]interface{}{
		"client_id":      client.ID,
		"session_id":     client.SessionID,
		"participant_id": client.ParticipantID,
		"is_host":        client.IsHost,
	})

	s.NotifyParticipantLeft(client.SessionID, client.ParticipantID)

	go func() {
		s.broadcastParticipantListUpdate(client.SessionID, "participant_disconnect", client)
	}()

	if client.IsHost {
		s.logger.Info("Host disconnected", map[string]interface{}{
			"session_id": client.SessionID,
		})

		s.stopQuestionTimer(client.SessionID)

		message := &models.WSMessage{
			Type: models.WSMsgTypeGamePaused,
			Payload: map[string]interface{}{
				"session_id": client.SessionID,
				"reason":     "host_disconnected",
				"paused_at":  time.Now(),
			},
			Timestamp: time.Now(),
		}

		s.broadcastToRoom(client.SessionID, message, client)
	}
}

// ===== Notification Methods for Broadcasting Events =====

func (s *gameEventHandler) NotifyGameStart(sessionID int64) error {
	message := &models.WSMessage{
		Type: models.WSMsgTypeQuizStart,
		Payload: map[string]interface{}{
			"session_id": sessionID,
			"started_at": time.Now(),
		},
		Timestamp: time.Now(),
	}

	return s.broadcastToRoom(sessionID, message, nil)
}

func (s *gameEventHandler) NotifyQuestionStart(sessionID int64, question *models.Question) error {
	// Get timer duration for client synchronization
	timeLimitSeconds := s.timeLimitToSeconds(question.TimeLimit)
	serverStartTime := time.Now()

	safeQuestion := map[string]interface{}{
		"id":         question.ID,
		"question":   question.Question,
		"type":       question.Type,
		"time_limit": timeLimitSeconds, // seconds
		"index":      question.Index,
		"max_score":  1000, // Maximum possible score
		"answers": func() []map[string]interface{} {
			answers := make([]map[string]interface{}, len(question.Answers))
			for i, answer := range question.Answers {
				answers[i] = map[string]interface{}{
					"text": answer.Text,
				}
			}
			return answers
		}(),
	}

	message := &models.WSMessage{
		Type: models.WSMsgTypeQuestionStart,
		Payload: map[string]interface{}{
			"session_id":        sessionID,
			"question":          safeQuestion,
			"started_at":        serverStartTime,
			"server_time_limit": timeLimitSeconds,
			"auto_advance":      true,
			"deadline":          serverStartTime.Add(time.Duration(timeLimitSeconds) * time.Second),
		},
		Timestamp: serverStartTime,
	}

	// Start automatic timer for question progression
	s.startQuestionTimer(sessionID, timeLimitSeconds)

	return s.broadcastToRoom(sessionID, message, nil)
}

func (s *gameEventHandler) NotifyQuestionEnd(sessionID int64) error {
	s.stopQuestionTimer(sessionID)

	message := &models.WSMessage{
		Type: models.WSMsgTypeQuestionEnd,
		Payload: map[string]interface{}{
			"session_id": sessionID,
			"ended_at":   time.Now(),
		},
		Timestamp: time.Now(),
	}

	return s.broadcastToRoom(sessionID, message, nil)
}

func (s *gameEventHandler) NotifyParticipantJoin(sessionID int64, participant *models.SessionParticipant) error {
	message := &models.WSMessage{
		Type: models.WSMsgTypeParticipantJoin,
		Payload: map[string]interface{}{
			"session_id":     sessionID,
			"participant_id": participant.ID,
			"nickname":       participant.Nickname,
			"joined_at":      time.Now(),
		},
		Timestamp: time.Now(),
	}

	return s.broadcastToRoom(sessionID, message, nil)
}

func (s *gameEventHandler) NotifyParticipantLeft(sessionID int64, participantID int64) error {
	message := &models.WSMessage{
		Type: models.WSMsgTypeParticipantLeft,
		Payload: map[string]interface{}{
			"session_id":     sessionID,
			"participant_id": participantID,
			"left_at":        time.Now(),
		},
		Timestamp: time.Now(),
	}

	return s.broadcastToRoom(sessionID, message, nil)
}

func (s *gameEventHandler) evaluateAnswer(question *models.Question, payload *models.WSAnswerPayload) bool {
	switch question.Type {
	case models.QuestionTypeSingleChoice:
		// Single Choice: Use AnswerValue (single string)
		if payload.AnswerValue == nil {
			s.logger.Warn("Single choice question requires answer_value", map[string]interface{}{
				"question_id": question.ID,
			})
			return false
		}

		submittedAnswer := *payload.AnswerValue
		for _, answerData := range question.Answers {
			if answerData.IsCorrect && answerData.Text == submittedAnswer {
				return true
			}
		}
		return false

	case models.QuestionTypeMultipleChoice:
		// Multiple Choice: Use AnswerValues (array of strings)
		if len(payload.AnswerValues) == 0 {
			s.logger.Warn("Multiple choice question requires answer_values array", map[string]interface{}{
				"question_id": question.ID,
			})
			return false
		}

		submittedAnswers := payload.AnswerValues

		var correctAnswers []string
		for _, answerData := range question.Answers {
			if answerData.IsCorrect {
				correctAnswers = append(correctAnswers, answerData.Text)
			}
		}

		if len(submittedAnswers) != len(correctAnswers) {
			s.logger.Info("Multiple choice answer count mismatch", map[string]interface{}{
				"question_id":       question.ID,
				"submitted_count":   len(submittedAnswers),
				"correct_count":     len(correctAnswers),
				"submitted_answers": submittedAnswers,
				"correct_answers":   correctAnswers,
			})
			return false
		}

		submittedSorted := make([]string, len(submittedAnswers))
		copy(submittedSorted, submittedAnswers)
		sort.Strings(submittedSorted)

		correctSorted := make([]string, len(correctAnswers))
		copy(correctSorted, correctAnswers)
		sort.Strings(correctSorted)

		for i, correctAnswer := range correctSorted {
			if i >= len(submittedSorted) || submittedSorted[i] != correctAnswer {
				s.logger.Info("Multiple choice answer mismatch", map[string]interface{}{
					"question_id": question.ID,
					"expected":    correctAnswer,
					"got":         submittedSorted[i],
					"position":    i,
				})
				return false
			}
		}
		return true

	case models.QuestionTypeTextInput:
		// Text Input: Use AnswerValue (single string), case-insensitive matching
		if payload.AnswerValue == nil {
			s.logger.Warn("Text input question requires answer_value", map[string]interface{}{
				"question_id": question.ID,
			})
			return false
		}

		submittedAnswer := *payload.AnswerValue
		submittedLower := strings.ToLower(strings.TrimSpace(submittedAnswer))

		for _, answerData := range question.Answers {
			if answerData.IsCorrect {
				correctLower := strings.ToLower(strings.TrimSpace(answerData.Text))
				if submittedLower == correctLower {
					return true
				}
			}
		}

		s.logger.Info("Text input answer did not match any correct answers", map[string]interface{}{
			"question_id":      question.ID,
			"submitted_answer": submittedAnswer,
		})
		return false

	default:
		s.logger.Error("Unknown question type", map[string]interface{}{
			"question_id":   question.ID,
			"question_type": question.Type,
		})
		return false
	}
}

func (s *gameEventHandler) parsePayload(payload interface{}, target interface{}) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(payloadBytes, target)
}

func (s *gameEventHandler) sendToClient(client *ws.Client, message *models.WSMessage) {
	msgBytes, err := json.Marshal(message)
	if err != nil {
		s.logger.Error("Failed to marshal WebSocket message", err)
		return
	}

	s.hub.SendToClient(client, msgBytes)
}

func (s *gameEventHandler) broadcastToRoom(sessionID int64, message *models.WSMessage, excludeClient *ws.Client) error {
	cacheKey := ""
	if message.Type == models.WSMsgTypeLeaderboard ||
		message.Type == models.WSMsgTypeQuestionEnd ||
		message.Type == models.WSMsgTypeQuizStart {
		cacheKey = fmt.Sprintf("%s_%d_%v", message.Type, sessionID, message.Timestamp.Unix())
	}

	var msgBytes []byte
	var err error

	if cacheKey != "" {
		if cached, ok := s.marshalCache.Load(cacheKey); ok {
			msgBytes = cached.([]byte)
		} else {
			msgBytes, err = json.Marshal(message)
			if err != nil {
				s.logger.Error("Failed to marshal WebSocket message", err)
				return exception.InternalError("MARSHAL_ERROR", "Failed to marshal message")
			}
			s.marshalCache.Store(cacheKey, msgBytes)

			// Clean cache after 30 seconds to prevent memory leaks
			time.AfterFunc(30*time.Second, func() {
				s.marshalCache.Delete(cacheKey)
			})
		}
	} else {
		msgBytes, err = json.Marshal(message)
		if err != nil {
			s.logger.Error("Failed to marshal WebSocket message", err)
			return exception.InternalError("MARSHAL_ERROR", "Failed to marshal message")
		}
	}

	s.hub.BroadcastToRoom(sessionID, msgBytes, excludeClient)
	return nil
}

func (s *gameEventHandler) formatLeaderboard(participants []*models.LeaderboardParticipant) []map[string]interface{} {
	leaderboard := make([]map[string]interface{}, 0, len(participants))
	for _, participant := range participants {
		leaderboard = append(leaderboard, map[string]interface{}{
			"rank":           participant.Rank,
			"participant_id": participant.ID,
			"nickname":       participant.Nickname,
			"score":          participant.Score,
			"is_host":        participant.IsHost,
		})
	}
	return leaderboard
}

func (s *gameEventHandler) sendError(client *ws.Client, code, message string) {
	errorMsg := &models.WSMessage{
		Type: models.WSMsgTypeError,
		Payload: map[string]interface{}{
			"code":    code,
			"message": message,
		},
		Timestamp: time.Now(),
	}

	s.sendToClient(client, errorMsg)
}

func (s *gameEventHandler) timeLimitToSeconds(timeLimit models.TimeLimitType) int32 {
	switch timeLimit {
	case models.TimeLimitType5:
		return 5
	case models.TimeLimitType10:
		return 10
	case models.TimeLimitType20:
		return 20
	case models.TimeLimitType45:
		return 45
	case models.TimeLimitType80:
		return 80
	default:
		return constants.QuestionTimeLimit
	}
}

func (s *gameEventHandler) startQuestionTimer(sessionID int64, timeLimit int32) {
	s.timerMutex.Lock()
	defer s.timerMutex.Unlock()

	if existingTimer, exists := s.questionTimers[sessionID]; exists {
		existingTimer.Stop()
	}

	timer := time.AfterFunc(time.Duration(timeLimit)*time.Second, func() {
		s.handleQuestionTimeout(sessionID)
	})

	s.questionTimers[sessionID] = timer
}

func (s *gameEventHandler) stopQuestionTimer(sessionID int64) {
	s.timerMutex.Lock()
	defer s.timerMutex.Unlock()

	if timer, exists := s.questionTimers[sessionID]; exists {
		timer.Stop()
		delete(s.questionTimers, sessionID)
	}
}

func (s *gameEventHandler) handleQuestionTimeout(sessionID int64) {
	ctx := context.Background()

	queries := []func(context.Context) (any, error){
		func(ctx context.Context) (any, error) {
			return s.sessionRepo.GetSessionByID(ctx, sessionID)
		},
		func(ctx context.Context) (any, error) {
			// get questions assuming session exists
			session, err := s.sessionRepo.GetSessionByID(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			return s.questionRepo.GetQuestionListByQuiz(ctx, session.QuizID)
		},
	}

	results, err := utils.RunQueriesParallel(ctx, queries)
	if err != nil {
		s.logger.Error("Failed to get session/questions for timeout", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	session := results[0].(*models.QuizSession)
	questions := results[1].([]*models.Question)

	// End current question
	s.NotifyQuestionEnd(sessionID)

	// Show leaderboard for 5 seconds
	s.showIntermediateLeaderboard(sessionID)

	// Check if this was the last question
	if session.CurrentQuestionIndex >= int32(len(questions))-1 {
		// This was the last question - end the game

		time.AfterFunc(5*time.Second, func() {
			s.autoEndGame(sessionID)
		})
	} else {
		// More questions available - auto advance after leaderboard
		time.AfterFunc(5*time.Second, func() {
			s.autoNextQuestion(sessionID)
		})
	}
}

func (s *gameEventHandler) showIntermediateLeaderboard(sessionID int64) {
	ctx := context.Background()

	leaderboard, err := s.sessionRepo.GetSessionLeaderboard(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get intermediate leaderboard", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	message := &models.WSMessage{
		Type: models.WSMsgTypeLeaderboard,
		Payload: map[string]interface{}{
			"session_id":    sessionID,
			"leaderboard":   s.formatLeaderboard(leaderboard),
			"display_time":  5, // seconds
			"next_action":   "auto_next_question",
			"server_driven": true, // Indicates server controls progression
			"updated_at":    time.Now(),
		},
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(sessionID, message, nil)
}

func (s *gameEventHandler) autoNextQuestion(sessionID int64) {
	ctx := context.Background()

	session, err := s.sessionRepo.GetSessionByID(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get session for auto next question", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	if session.Status != models.SessionStatusActive {
		s.logger.Warn("Session not active, skipping auto next question", map[string]interface{}{
			"session_id": sessionID,
			"status":     session.Status,
		})
		return
	}

	nextQuestionIndex := session.CurrentQuestionIndex + 1

	questions, err := s.questionRepo.GetQuestionListByQuiz(ctx, session.QuizID)
	if err != nil {
		s.logger.Error("Failed to get questions for auto next", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	if nextQuestionIndex >= int32(len(questions)) {
		s.autoEndGame(sessionID)
		return
	}

	err = s.sessionRepo.UpdateSessionQuestion(ctx, sessionID, nextQuestionIndex)
	if err != nil {
		s.logger.Error("Failed to update session question for auto next", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	// Get the new question
	currentQuestion := questions[nextQuestionIndex]

	s.NotifyQuestionStart(sessionID, currentQuestion)

	// Start timer for this question based on its time limit
	timeLimitSeconds := s.timeLimitToSeconds(currentQuestion.TimeLimit)
	s.startQuestionTimer(sessionID, timeLimitSeconds)
}

func (s *gameEventHandler) autoEndGame(sessionID int64) {
	ctx := context.Background()

	err := s.sessionRepo.EndSession(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to auto-end session", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		return
	}

	// Get final leaderboard
	finalLeaderboard, err := s.sessionRepo.GetSessionLeaderboard(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get final leaderboard for auto-end", map[string]interface{}{
			"session_id": sessionID,
			"error":      err.Error(),
		})
		finalLeaderboard = []*models.LeaderboardParticipant{}
	}

	message := &models.WSMessage{
		Type: models.WSMsgTypeQuizEnd,
		Payload: map[string]interface{}{
			"session_id":        sessionID,
			"ended_at":          time.Now(),
			"final_leaderboard": s.formatLeaderboard(finalLeaderboard),
			"status":            "completed",
			"auto_ended":        true,
			"completion_reason": "all_questions_completed",
			"server_driven":     true,
		},
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(sessionID, message, nil)

	// Clean up timer
	s.stopQuestionTimer(sessionID)
}

func (s *gameEventHandler) broadcastParticipantListUpdate(sessionID int64, trigger string, excludeClient *ws.Client) {
	ctx := context.Background()
	participants, err := s.sessionRepo.GetSessionParticipants(ctx, sessionID)
	if err != nil {
		s.logger.Error("Failed to get participants for list update", map[string]interface{}{
			"session_id": sessionID,
			"trigger":    trigger,
			"error":      err.Error(),
		})
		return
	}

	participantList := make([]map[string]interface{}, len(participants))
	for i, p := range participants {
		participantList[i] = map[string]interface{}{
			"id":       p.ID,
			"nickname": p.Nickname,
			"is_host":  p.IsHost,
			"score":    p.Score,
		}
	}

	message := &models.WSMessage{
		Type: models.WSMsgTypeParticipantListUpdate,
		Payload: map[string]interface{}{
			"session_id":        sessionID,
			"participant_count": len(participants),
			"participants":      participantList,
			"trigger":           trigger,
			"updated_at":        time.Now(),
		},
		Timestamp: time.Now(),
	}

	s.broadcastToRoom(sessionID, message, excludeClient)
}
