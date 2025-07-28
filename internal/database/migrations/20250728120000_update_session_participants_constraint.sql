-- +goose Up
-- +goose StatementBegin

-- Drop the old constraint
DROP INDEX IF EXISTS idx_session_participants_unique_authenticated_user;

-- Create new constraint that only applies to real user IDs (1 to 999,999,999)
-- Anonymous users use IDs >= 1,000,000,000, so they can have duplicates
CREATE UNIQUE INDEX idx_session_participants_unique_authenticated_user 
ON session_participants(session_id, user_id) 
WHERE user_id > 0 AND user_id < 1000000000;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to the old constraint
DROP INDEX IF EXISTS idx_session_participants_unique_authenticated_user;

CREATE UNIQUE INDEX idx_session_participants_unique_authenticated_user 
ON session_participants(session_id, user_id) 
WHERE user_id > 0;

-- +goose StatementEnd
