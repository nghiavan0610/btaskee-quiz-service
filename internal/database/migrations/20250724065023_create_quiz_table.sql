-- +goose Up
-- +goose StatementBegin
CREATE TYPE question_type AS ENUM ('single_choice', 'multiple_choice', 'text_input');
CREATE TYPE time_limit_type AS ENUM ('5', '10', '20', '45', '80');
CREATE TYPE quiz_visibility AS ENUM ('private', 'published');

CREATE TABLE IF NOT EXISTS quizzes (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    visibility quiz_visibility NOT NULL DEFAULT 'private',
    slug VARCHAR(100) UNIQUE,
    view_count INTEGER NOT NULL DEFAULT 0,
    play_count INTEGER NOT NULL DEFAULT 0,
    max_participants INTEGER DEFAULT 100,
    current_question_index INTEGER NOT NULL DEFAULT 0,
    total_questions INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    published_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS questions (
    id BIGSERIAL PRIMARY KEY,
    quiz_id BIGINT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question VARCHAR(100) NOT NULL,
    type question_type NOT NULL DEFAULT 'single_choice',
    answers JSONB DEFAULT '[]',
    time_limit time_limit_type NOT NULL DEFAULT '20', -- seconds
    index INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);


-- Quizzes table indexes
CREATE INDEX idx_quizzes_created_at ON quizzes(created_at DESC);
CREATE INDEX idx_quizzes_visibility ON quizzes(visibility);
CREATE INDEX idx_quizzes_published ON quizzes(published_at DESC) WHERE published_at IS NOT NULL;
CREATE INDEX idx_quizzes_slug ON quizzes(slug) WHERE slug IS NOT NULL;

-- Questions table indexes
CREATE INDEX idx_questions_quiz_id ON questions(quiz_id);
CREATE INDEX idx_questions_index ON questions(quiz_id, index);
CREATE INDEX idx_questions_type_time ON questions(type, time_limit);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_questions_type_time;
DROP INDEX IF EXISTS idx_questions_order;
DROP INDEX IF EXISTS idx_questions_quiz_id;
DROP INDEX IF EXISTS idx_quizzes_slug;
DROP INDEX IF EXISTS idx_quizzes_published;
DROP INDEX IF EXISTS idx_quizzes_public;
DROP INDEX IF EXISTS idx_quizzes_visibility;
DROP INDEX IF EXISTS idx_quizzes_created_at;
DROP INDEX IF EXISTS idx_quizzes_status;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS questions;
DROP TABLE IF EXISTS quizzes;

-- Drop types
DROP TYPE IF EXISTS quiz_visibility;
DROP TYPE IF EXISTS time_limit_type;
DROP TYPE IF EXISTS question_type;
-- +goose StatementEnd
