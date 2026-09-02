-- +goose Up
-- +goose StatementBegin
CREATE TABLE api_tokens (
    id         SERIAL PRIMARY KEY,
    user_id    INT REFERENCES users (id) ON DELETE CASCADE,
    gallery_id INT REFERENCES galleries (id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE api_tokens;
-- +goose StatementEnd
