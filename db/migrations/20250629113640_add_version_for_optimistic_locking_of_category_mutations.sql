-- +goose Up
-- +goose StatementBegin
ALTER TABLE categories ADD COLUMN version BIGINT NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE categories DROP COLUMN version;
-- +goose StatementEnd
