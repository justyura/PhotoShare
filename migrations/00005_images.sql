-- +goose Up
-- +goose StatementBegin
CREATE TABLE images (
    id         SERIAL PRIMARY KEY,
    uuid       TEXT UNIQUE NOT NULL,
    gallery_id INT REFERENCES galleries (id) ON DELETE CASCADE,
    ext        TEXT NOT NULL,
    orig_name  TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE images;
-- +goose StatementEnd
