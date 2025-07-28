-- name: CreateQuestion :one
INSERT INTO questions (quiz_id, question, type, answers, index, time_limit)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetQuestionByID :one
SELECT * FROM questions WHERE id = $1;

-- name: GetMaxQuestionIndexByQuiz :one
SELECT COALESCE(MAX(index), 0)::int FROM questions WHERE quiz_id = $1;

-- name: GetQuestionListByQuiz :many
SELECT * FROM questions 
WHERE quiz_id = $1 
ORDER BY index ASC;

-- name: GetQuestionByQuizAndIndex :one
SELECT * FROM questions 
WHERE quiz_id = $1 AND index = $2;

-- name: GetCurrentQuestion :one
SELECT q.* FROM questions q
JOIN quizzes qz ON q.quiz_id = qz.id
WHERE qz.id = $1 AND q.index = qz.current_question_index;

-- name: UpdateQuestion :one
UPDATE questions 
SET question = $2, type = $3, answers = $4, time_limit = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteQuestion :exec
DELETE FROM questions WHERE id = $1;

-- name: CountQuestionsByQuiz :one
SELECT COUNT(*) FROM questions WHERE quiz_id = $1;

-- name: UpdateQuestionIndex :exec
UPDATE questions 
SET index = $2, updated_at = NOW()
WHERE id = $1;
