-- +goose Up
-- +goose StatementBegin
ALTER TABLE base_references ADD COLUMN is_starred BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE base_references DROP COLUMN is_starred;
-- +goose StatementEnd