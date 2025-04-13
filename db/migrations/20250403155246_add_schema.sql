-- +goose Up
-- +goose StatementBegin

CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL,
    position INTEGER NOT NULL
);

CREATE TABLE base_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    position INTEGER NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);

CREATE TABLE book_references (
    reference_id INTEGER PRIMARY KEY,
    isbn VARCHAR(50) NOT NULL,
    FOREIGN KEY (reference_id) REFERENCES base_references(id) ON DELETE CASCADE
);

CREATE TABLE link_references (
    reference_id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY (reference_id) REFERENCES base_references(id) ON DELETE CASCADE
);

CREATE TABLE note_references (
    reference_id INTEGER PRIMARY KEY,
    text TEXT NOT NULL,
    FOREIGN KEY (reference_id) REFERENCES base_references(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE note_references;
DROP TABLE link_references;
DROP TABLE book_references;
DROP TABLE base_references;
DROP TABLE categories;

-- +goose StatementEnd
