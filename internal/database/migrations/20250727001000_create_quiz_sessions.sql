-- +goose Up
-- +goose StatementBegin

-- Quiz session status enum
CREATE TYPE session_status AS ENUM ('waiting', 'active', 'completed', 'cancelled');

-- Quiz sessions table (ESSENTIAL - for managing quiz sessions)
CREATE TABLE IF NOT EXISTS quiz_sessions (
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    host_id BIGINT, -- Can be NULL for guest hosts, or generated anonymous ID (no foreign key constraint)
    join_code VARCHAR(8) UNIQUE NOT NULL,
    status session_status NOT NULL DEFAULT 'waiting',
    current_question_index INTEGER NOT NULL DEFAULT 0,
    max_participants INTEGER DEFAULT 100,
    participant_count INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Session participants table (ESSENTIAL - for real-time scoring and leaderboard)
CREATE TABLE IF NOT EXISTS session_participants (
    id BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES quiz_sessions(id) ON DELETE CASCADE,
    user_id BIGINT, -- Can be NULL for guest participants, or generated anonymous ID (no foreign key constraint)
    nickname VARCHAR(50) NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,  -- THIS IS ALL YOU NEED for real-time leaderboard!
    is_host BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    -- Note: No UNIQUE constraint here to allow multiple anonymous users
);

-- Essential indexes for performance
CREATE INDEX idx_quiz_sessions_join_code ON quiz_sessions(join_code);
CREATE INDEX idx_quiz_sessions_status ON quiz_sessions(status);
CREATE INDEX idx_quiz_sessions_quiz_id ON quiz_sessions(quiz_id);

-- Critical index for fast real-time leaderboard queries
CREATE INDEX idx_session_participants_session_id ON session_participants(session_id);
CREATE INDEX idx_session_participants_score ON session_participants(session_id, score DESC);

-- Partial unique index: prevent duplicate authenticated user joins, but allow multiple anonymous users
-- We assume user IDs from the users table start from 1, so negative IDs are for anonymous users
CREATE UNIQUE INDEX idx_session_participants_unique_authenticated_user 
ON session_participants(session_id, user_id) 
WHERE user_id > 0;

-- Function to update participant count (keeps sessions table in sync)
CREATE OR REPLACE FUNCTION update_participant_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE quiz_sessions 
        SET participant_count = participant_count + 1 
        WHERE id = NEW.session_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE quiz_sessions 
        SET participant_count = participant_count - 1 
        WHERE id = OLD.session_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update participant count
CREATE TRIGGER trigger_update_participant_count
    AFTER INSERT OR DELETE ON session_participants
    FOR EACH ROW EXECUTE FUNCTION update_participant_count();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_update_participant_count ON session_participants;

-- Drop functions
DROP FUNCTION IF EXISTS update_participant_count();

-- Drop constraints and indexes
DROP INDEX IF EXISTS idx_session_participants_unique_authenticated_user;
DROP INDEX IF EXISTS idx_session_participants_score;
DROP INDEX IF EXISTS idx_session_participants_session_id;
DROP INDEX IF EXISTS idx_quiz_sessions_quiz_id;
DROP INDEX IF EXISTS idx_quiz_sessions_status;
DROP INDEX IF EXISTS idx_quiz_sessions_join_code;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS session_participants;
DROP TABLE IF EXISTS quiz_sessions;

-- Drop types
DROP TYPE IF EXISTS session_status;

-- +goose StatementEnd
