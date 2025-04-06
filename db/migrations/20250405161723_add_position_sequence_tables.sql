-- +goose Up
-- +goose StatementBegin
CREATE TABLE category_position_sequences (
    id INTEGER PRIMARY KEY,
    next_position INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE reference_position_sequences (
    category_id INTEGER NOT NULL,
    next_position INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (category_id)
);

INSERT INTO category_position_sequence (id, next_position) VALUES (1, 0);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reference_position_sequences;
DROP TABLE category_position_sequences;
-- +goose StatementEnd
