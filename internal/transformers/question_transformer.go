package transformers

import (
	"encoding/json"
	"fmt"

	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
)

func ConvertToQuestionModel(result sqlc.Question) (*models.Question, error) {
	question := &models.Question{
		ID:        result.ID,
		QuizID:    result.QuizID,
		Question:  result.Question,
		Type:      models.QuestionType(result.Type),
		Index:     result.Index,
		TimeLimit: models.TimeLimitType(result.TimeLimit),
		CreatedAt: result.CreatedAt.Time,
		UpdatedAt: result.UpdatedAt.Time,
	}

	if len(result.Answers) > 0 {
		answers, err := ParseAnswersFromJSON(result.Answers)
		if err != nil {
			return nil, fmt.Errorf("failed to parse answers JSON: %w", err)
		}
		question.Answers = answers
	} else {
		question.Answers = []models.AnswerData{}
	}

	return question, nil
}

func ConvertAnswersToJSON(answers []models.AnswerData) ([]byte, error) {
	if answers == nil {
		return json.Marshal([]models.AnswerData{})
	}

	return json.Marshal(answers)
}

func ParseAnswersFromJSON(jsonData []byte) ([]models.AnswerData, error) {
	if len(jsonData) == 0 {
		return []models.AnswerData{}, nil
	}

	var answers []models.AnswerData
	if err := json.Unmarshal(jsonData, &answers); err != nil {
		return nil, fmt.Errorf("failed to parse answers JSON: %w", err)
	}

	return answers, nil
}

func ValidateAnswersFormat(answers []models.AnswerData, questionType models.QuestionType) error {
	switch questionType {
	case models.QuestionTypeSingleChoice:
		// Should have at least 2 options, exactly 1 correct
		if len(answers) < 2 {
			return fmt.Errorf("single choice questions must have at least 2 options")
		}
		correctCount := 0
		for _, answer := range answers {
			if answer.IsCorrect {
				correctCount++
			}
		}
		if correctCount != 1 {
			return fmt.Errorf("single choice questions must have exactly 1 correct answer")
		}

	case models.QuestionTypeMultipleChoice:
		// Should have at least 2 options, at least 1 correct
		if len(answers) < 2 {
			return fmt.Errorf("multiple choice questions must have at least 2 options")
		}
		correctCount := 0
		for _, answer := range answers {
			if answer.IsCorrect {
				correctCount++
			}
		}
		if correctCount == 0 {
			return fmt.Errorf("multiple choice questions must have at least 1 correct answer")
		}

	case models.QuestionTypeTextInput:
	}

	return nil
}
