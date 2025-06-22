-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_categories_position_unique ON categories(position);
CREATE UNIQUE INDEX idx_base_references_category_position_unique ON base_references(category_id, position);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_base_references_category_position_unique;
DROP INDEX IF EXISTS idx_categories_position_unique;
-- +goose StatementEnd
