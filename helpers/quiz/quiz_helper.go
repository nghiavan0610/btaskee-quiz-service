package helpers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
)

type QuizCursorData struct {
	ID        int64      `json:"id"`
	Title     *string    `json:"title,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	ViewCount *int32     `json:"view_count,omitempty"`
	PlayCount *int32     `json:"play_count,omitempty"`
	SortBy    string     `json:"sort_by"`
}

func EncodeQuizCursor(quiz *models.Quiz, sortBy string) (string, error) {
	if quiz == nil {
		return "", fmt.Errorf("quiz cannot be nil")
	}

	cursor := QuizCursorData{
		ID:     quiz.ID,
		SortBy: sortBy,
	}

	switch sortBy {
	case "name_asc", "name_desc":
		cursor.Title = &quiz.Title
	case "time_newest", "time_oldest":
		cursor.CreatedAt = &quiz.CreatedAt
	case "view_count":
		cursor.ViewCount = &quiz.ViewCount
	case "play_count":
		cursor.PlayCount = &quiz.PlayCount
	default:
		// Default to time-based cursor
		cursor.CreatedAt = &quiz.CreatedAt
	}

	jsonData, err := json.Marshal(cursor)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor data: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(jsonData)
	return encoded, nil
}

func DecodeQuizCursor(cursorStr string) (*QuizCursorData, error) {
	if cursorStr == "" {
		return nil, nil
	}

	decodedBytes, err := base64.URLEncoding.DecodeString(cursorStr)
	if err == nil {
		var cursor QuizCursorData
		if jsonErr := json.Unmarshal(decodedBytes, &cursor); jsonErr == nil {
			return &cursor, nil
		}
	}

	if id, parseErr := strconv.ParseInt(cursorStr, 10, 64); parseErr == nil {
		return &QuizCursorData{
			ID:     id,
			SortBy: "time_newest", // Default
		}, nil
	}

	return nil, fmt.Errorf("invalid cursor format")
}

func ValidateQuizSortBy(sortBy *string) string {
	if sortBy == nil || *sortBy == "" {
		return "time_newest"
	}

	validSorts := map[string]bool{
		"name_asc":    true,
		"name_desc":   true,
		"time_newest": true,
		"time_oldest": true,
		"view_count":  true,
		"play_count":  true,
	}

	if validSorts[*sortBy] {
		return *sortBy
	}

	return "time_newest"
}
