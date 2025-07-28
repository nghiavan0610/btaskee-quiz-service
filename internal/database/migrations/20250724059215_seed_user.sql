-- +goose Up
-- +goose StatementBegin
INSERT INTO users (username, email, password, is_active) VALUES
('ben', 'ben@yopmail.com', 'abc@123X', true);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE id IN (1);

-- +goose StatementEnd
