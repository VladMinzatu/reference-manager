-- +goose Up
-- +goose StatementBegin

-- We keep positions in separate tables because:
-- 1. It allows us to update positions without locking the main tables
-- 2. It separates concerns - positions are about presentation order, not core data
CREATE TABLE categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE category_positions (
    category_id INTEGER PRIMARY KEY,
    position INTEGER NOT NULL,
    UNIQUE(position),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE TABLE base_references (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE TABLE book_references (
    reference_id INTEGER PRIMARY KEY,
    isbn VARCHAR(50) NOT NULL,
    FOREIGN KEY (reference_id) REFERENCES base_references(id)
);

CREATE TABLE link_references (
    reference_id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY (reference_id) REFERENCES base_references(id)
);

CREATE TABLE note_references (
    reference_id INTEGER PRIMARY KEY,
    text TEXT NOT NULL,
    FOREIGN KEY (reference_id) REFERENCES base_references(id)
);

CREATE TABLE reference_positions (
    reference_id INTEGER PRIMARY KEY,
    category_id INTEGER NOT NULL,
    position INTEGER NOT NULL,
    UNIQUE(category_id, position),
    FOREIGN KEY (reference_id) REFERENCES base_references(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE reference_positions;
DROP TABLE note_references;
DROP TABLE link_references;
DROP TABLE book_references;
DROP TABLE base_references;
DROP TABLE category_positions;
DROP TABLE categories;

-- +goose StatementEnd
