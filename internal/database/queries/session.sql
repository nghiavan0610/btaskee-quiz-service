-- name: CreateSession :one
INSERT INTO quiz_sessions (
    quiz_id, host_id, join_code, status, max_participants,
    current_question_index, participant_count
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSessionByID :one
SELECT 
    s.*,
    q.title as quiz_title,
    q.description as quiz_description,
    q.total_questions as quiz_total_questions
FROM quiz_sessions s
JOIN quizzes q ON s.quiz_id = q.id
WHERE s.id = $1;

-- name: GetSessionByJoinCode :one
SELECT 
    s.*,
    q.title as quiz_title,
    q.description as quiz_description,
    q.total_questions as quiz_total_questions
FROM quiz_sessions s
JOIN quizzes q ON s.quiz_id = q.id
WHERE s.join_code = $1;

-- name: UpdateSession :one
UPDATE quiz_sessions 
SET 
    status = $2,
    current_question_index = $3,
    max_participants = $4,
    participant_count = $5,
    started_at = $6,
    ended_at = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: StartSession :exec
UPDATE quiz_sessions 
SET 
    status = 'active',
    started_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND status = 'waiting';

-- name: EndSession :exec
UPDATE quiz_sessions 
SET 
    status = 'completed',
    ended_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND status = 'active';

-- name: UpdateSessionQuestion :exec
UPDATE quiz_sessions 
SET 
    current_question_index = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: AddParticipant :one
INSERT INTO session_participants (
    session_id, user_id, nickname, score, is_host
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetSessionParticipants :many
SELECT * FROM session_participants
WHERE session_id = $1
ORDER BY joined_at ASC;

-- name: GetSessionLeaderboard :many
SELECT 
    id,
    nickname,
    score,
    is_host,
    ROW_NUMBER() OVER (ORDER BY score DESC, joined_at ASC) as rank
FROM session_participants
WHERE session_id = $1
ORDER BY score DESC, joined_at ASC;

-- name: UpdateParticipantScore :exec
UPDATE session_participants 
SET 
    score = score + $2,  -- Add to existing score for real-time updates
    last_activity = NOW()
WHERE id = $1;

-- name: CheckJoinCodeExists :one
SELECT EXISTS(SELECT 1 FROM quiz_sessions WHERE join_code = $1);
