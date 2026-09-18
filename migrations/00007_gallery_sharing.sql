-- +goose Up
-- +goose StatementBegin
ALTER TABLE galleries
    ADD COLUMN is_public BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN share_id TEXT UNIQUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE galleries
    DROP COLUMN share_id,
    DROP COLUMN is_public;
-- +goose StatementEnd
