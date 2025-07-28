package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	helpers "github.com/nghiavan0610/btaskee-quiz-service/helpers/quiz"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/models"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/transformers"
	"github.com/samber/lo"
)

type (
	QuizRepository interface {
		CreateQuiz(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error)
		UpdateQuiz(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error)
		GetQuizDetail(ctx context.Context, id int64, includeQuestions bool) (*models.Quiz, error)
		DeleteQuiz(ctx context.Context, id int64) error
		GetQuizListByOwner(ctx context.Context, ownerID int64, query *string, visibility *models.QuizVisibility, limit, offset int32) ([]*models.Quiz, error)
		CountQuizListByOwner(ctx context.Context, ownerID int64, query *string, visibility *models.QuizVisibility) (int64, error)
		GetPublicQuizList(ctx context.Context, query *string, cursor *string, limit int32, sortBy *string) ([]*models.Quiz, error)
		IncrementViewCount(ctx context.Context, quizID int64) error
		IncrementPlayCount(ctx context.Context, quizID int64) error
		UpdateTotalQuestions(ctx context.Context, quizID int64, totalQuestions int32) error
	}

	quizRepository struct {
		queries *sqlc.Queries
	}
)

func ProvideQuizRepository(queries *sqlc.Queries) QuizRepository {
	return &quizRepository{
		queries: queries,
	}
}

func (r *quizRepository) CreateQuiz(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error) {
	params := sqlc.CreateQuizParams{
		Title:           quiz.Title,
		Description:     quiz.Description,
		Visibility:      sqlc.QuizVisibilityPrivate,
		OwnerID:         quiz.OwnerID,
		MaxParticipants: quiz.MaxParticipants,
	}

	result, err := r.queries.CreateQuiz(ctx, params)
	if err != nil {
		return nil, err
	}

	return transformers.ConvertToQuizModel(result), nil
}

func (r *quizRepository) GetQuizDetail(ctx context.Context, id int64, includeQuestions bool) (*models.Quiz, error) {
	result, err := r.queries.GetQuizWithOwner(ctx, id)
	if err != nil {
		return nil, err
	}

	quiz := transformers.ConvertToQuizModel(result)

	// Load questions if requested
	if includeQuestions {
		questions, err := r.queries.GetQuestionListByQuiz(ctx, id)
		if err != nil {
			return nil, err
		}

		quiz.Questions = make([]models.Question, len(questions))
		for i, q := range questions {
			answers, err := transformers.ParseAnswersFromJSON(q.Answers)
			if err != nil {
				return nil, err
			}

			quiz.Questions[i] = models.Question{
				ID:        q.ID,
				QuizID:    q.QuizID,
				Question:  q.Question,
				Type:      models.QuestionType(q.Type),
				Answers:   answers,
				TimeLimit: models.TimeLimitType(q.TimeLimit),
				Index:     q.Index,
				CreatedAt: q.CreatedAt.Time,
				UpdatedAt: q.UpdatedAt.Time,
			}
		}
	}

	return quiz, nil
}

func (r *quizRepository) UpdateQuiz(ctx context.Context, quiz *models.Quiz) (*models.Quiz, error) {
	var publishedAt pgtype.Timestamptz
	if quiz.PublishedAt != nil {
		publishedAt = pgtype.Timestamptz{
			Time:  *quiz.PublishedAt,
			Valid: true,
		}
	}

	result, err := r.queries.UpdateQuiz(ctx, sqlc.UpdateQuizParams{
		ID:              quiz.ID,
		Title:           quiz.Title,
		Description:     quiz.Description,
		Visibility:      sqlc.QuizVisibility(quiz.Visibility),
		MaxParticipants: quiz.MaxParticipants,
		PublishedAt:     publishedAt,
	})
	if err != nil {
		return nil, err
	}

	return transformers.ConvertToQuizModel(result), nil
}

func (r *quizRepository) DeleteQuiz(ctx context.Context, id int64) error {
	err := r.queries.DeleteQuiz(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

func (r *quizRepository) GetQuizListByOwner(ctx context.Context, ownerID int64, query *string, visibility *models.QuizVisibility, limit, offset int32) ([]*models.Quiz, error) {
	var visibilityParam sqlc.NullQuizVisibility
	if lo.FromPtr(visibility) != "" {
		visibilityParam = sqlc.NullQuizVisibility{
			QuizVisibility: sqlc.QuizVisibility(*visibility),
			Valid:          true,
		}
	} else {
		visibilityParam = sqlc.NullQuizVisibility{Valid: false}
	}

	results, err := r.queries.GetQuizListByOwner(ctx, sqlc.GetQuizListByOwnerParams{
		OwnerID:    ownerID,
		Query:      query,
		Visibility: visibilityParam,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	quizzes := make([]*models.Quiz, len(results))
	for i, result := range results {
		quizzes[i] = transformers.ConvertToQuizModel(result)
	}

	return quizzes, nil
}

func (r *quizRepository) CountQuizListByOwner(ctx context.Context, ownerID int64, query *string, visibility *models.QuizVisibility) (int64, error) {
	var visibilityParam sqlc.NullQuizVisibility
	if lo.FromPtr(visibility) != "" {
		visibilityParam = sqlc.NullQuizVisibility{
			QuizVisibility: sqlc.QuizVisibility(*visibility),
			Valid:          true,
		}
	} else {
		visibilityParam = sqlc.NullQuizVisibility{Valid: false}
	}

	count, err := r.queries.CountQuizListByOwner(ctx, sqlc.CountQuizListByOwnerParams{
		OwnerID:    ownerID,
		Query:      query,
		Visibility: visibilityParam,
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *quizRepository) GetPublicQuizList(ctx context.Context, query *string, cursor *string, limit int32, sortBy *string) ([]*models.Quiz, error) {
	validSortBy := helpers.ValidateQuizSortBy(sortBy)

	// Decode cursor if provided
	cursorData, err := helpers.DecodeQuizCursor(lo.FromPtr(cursor))
	if err != nil {
		return nil, err
	}

	params := sqlc.GetPublicQuizzesParams{
		Query:  query,
		Limit:  limit,
		SortBy: &validSortBy,
	}

	if cursorData != nil {
		params.CursorID = &cursorData.ID

		switch validSortBy {
		case "name_asc", "name_desc":
			if cursorData.Title != nil {
				params.CursorTitle = cursorData.Title
			}
		case "time_newest", "time_oldest":
			if cursorData.CreatedAt != nil {
				params.CursorCreatedAt = pgtype.Timestamptz{
					Time:  *cursorData.CreatedAt,
					Valid: true,
				}
			}
		case "view_count":
			if cursorData.ViewCount != nil {
				params.CursorViewCount = cursorData.ViewCount
			}
		case "play_count":
			if cursorData.PlayCount != nil {
				params.CursorPlayCount = cursorData.PlayCount
			}
		}
	}

	quizzes, err := r.queries.GetPublicQuizzes(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*models.Quiz, len(quizzes))
	for i, quiz := range quizzes {
		result[i] = transformers.ConvertToQuizModel(quiz)
	}

	return result, nil
}

func (r *quizRepository) IncrementViewCount(ctx context.Context, quizID int64) error {
	return r.queries.IncrementQuizViewCount(ctx, quizID)
}

func (r *quizRepository) IncrementPlayCount(ctx context.Context, quizID int64) error {
	return r.queries.IncrementQuizPlayCount(ctx, quizID)
}

func (r *quizRepository) UpdateTotalQuestions(ctx context.Context, quizID int64, totalQuestions int32) error {
	params := sqlc.UpdateQuizTotalQuestionsParams{
		ID:             quizID,
		TotalQuestions: &totalQuestions,
	}
	return r.queries.UpdateQuizTotalQuestions(ctx, params)
}
