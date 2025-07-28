-- name: CreateQuiz :one
INSERT INTO quizzes (title, description, visibility, owner_id, max_participants)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, description, owner_id, visibility, slug, view_count, play_count, max_participants, current_question_index, total_questions, created_at, updated_at, published_at;

-- name: UpdateQuiz :one
UPDATE quizzes 
SET title = $2, description = $3, visibility = $4, max_participants = $5, published_at = $6, updated_at = NOW()
WHERE id = $1
RETURNING id, title, description, owner_id, visibility, slug, view_count, play_count, max_participants, current_question_index, total_questions, created_at, updated_at, published_at;

-- name: GetQuizWithOwner :one
SELECT 
    q.*,
    u.id as owner_user_id,
    u.username as owner_username,
    u.email as owner_email,
    u.avatar_url as owner_avatar_url
FROM quizzes q
LEFT JOIN users u ON q.owner_id = u.id
WHERE q.id = $1;

-- name: GetQuizListByOwner :many
SELECT id, title, description, owner_id, visibility, slug, view_count, play_count, max_participants, current_question_index, total_questions, created_at, updated_at, published_at FROM quizzes 
WHERE owner_id = $1
  AND (sqlc.narg('query')::text IS NULL OR (title ILIKE '%' || sqlc.narg('query') || '%' OR description ILIKE '%' || sqlc.narg('query') || '%'))
  AND (sqlc.narg('visibility')::quiz_visibility IS NULL OR visibility = sqlc.narg('visibility'))
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountQuizListByOwner :one
SELECT COUNT(*) FROM quizzes
WHERE owner_id = $1
  AND (sqlc.narg('query')::text IS NULL OR (title ILIKE '%' || sqlc.narg('query') || '%' OR description ILIKE '%' || sqlc.narg('query') || '%'))
  AND (sqlc.narg('visibility')::quiz_visibility IS NULL OR visibility = sqlc.narg('visibility')); 

-- name: DeleteQuiz :exec
DELETE FROM quizzes WHERE id = $1;

-- name: GetPublicQuizzes :many
SELECT 
    q.id, q.title, q.description, q.owner_id, q.visibility, q.slug, q.view_count, q.play_count, q.max_participants, q.current_question_index, q.total_questions, q.created_at, q.updated_at, q.published_at,
    u.id as owner_user_id,
    u.username as owner_username,
    u.email as owner_email,
    u.avatar_url as owner_avatar_url
FROM quizzes q
LEFT JOIN users u ON q.owner_id = u.id
WHERE q.visibility = 'published'
  AND (sqlc.narg('query')::text IS NULL OR (q.title ILIKE '%' || sqlc.narg('query') || '%' OR q.description ILIKE '%' || sqlc.narg('query') || '%'))
  AND (
    sqlc.narg('cursor_id')::bigint IS NULL OR 
    CASE sqlc.narg('sort_by')::text
      WHEN 'name_asc' THEN q.title > sqlc.narg('cursor_title')::text OR (q.title = sqlc.narg('cursor_title')::text AND q.id > sqlc.narg('cursor_id')::bigint)
      WHEN 'name_desc' THEN q.title < sqlc.narg('cursor_title')::text OR (q.title = sqlc.narg('cursor_title')::text AND q.id > sqlc.narg('cursor_id')::bigint)
      WHEN 'time_newest' THEN q.created_at < sqlc.narg('cursor_created_at')::timestamptz OR (q.created_at = sqlc.narg('cursor_created_at')::timestamptz AND q.id > sqlc.narg('cursor_id')::bigint)
      WHEN 'time_oldest' THEN q.created_at > sqlc.narg('cursor_created_at')::timestamptz OR (q.created_at = sqlc.narg('cursor_created_at')::timestamptz AND q.id > sqlc.narg('cursor_id')::bigint)
      WHEN 'view_count' THEN q.view_count < sqlc.narg('cursor_view_count')::integer OR (q.view_count = sqlc.narg('cursor_view_count')::integer AND q.id > sqlc.narg('cursor_id')::bigint)
      WHEN 'play_count' THEN q.play_count < sqlc.narg('cursor_play_count')::integer OR (q.play_count = sqlc.narg('cursor_play_count')::integer AND q.id > sqlc.narg('cursor_id')::bigint)
      ELSE q.created_at < sqlc.narg('cursor_created_at')::timestamptz OR (q.created_at = sqlc.narg('cursor_created_at')::timestamptz AND q.id > sqlc.narg('cursor_id')::bigint)
    END
  )
ORDER BY 
  CASE sqlc.narg('sort_by')::text
    WHEN 'name_asc' THEN q.title
  END ASC,
  CASE sqlc.narg('sort_by')::text
    WHEN 'name_desc' THEN q.title
  END DESC,
  CASE sqlc.narg('sort_by')::text
    WHEN 'time_newest' THEN q.created_at
  END DESC,
  CASE sqlc.narg('sort_by')::text
    WHEN 'time_oldest' THEN q.created_at
  END ASC,
  CASE sqlc.narg('sort_by')::text
    WHEN 'view_count' THEN q.view_count
  END DESC,
  CASE sqlc.narg('sort_by')::text
    WHEN 'play_count' THEN q.play_count
  END DESC,
  CASE 
    WHEN sqlc.narg('sort_by')::text IS NULL OR sqlc.narg('sort_by')::text = '' THEN q.created_at
  END DESC,
  q.id ASC
LIMIT sqlc.arg('limit')::integer;

-- name: IncrementQuizViewCount :exec
UPDATE quizzes 
SET view_count = view_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: IncrementQuizPlayCount :exec
UPDATE quizzes 
SET play_count = play_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: UpdateQuizTotalQuestions :exec
UPDATE quizzes 
SET total_questions = $2, updated_at = NOW()
WHERE id = $1;