-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE category_positions (
    category_id BIGINT PRIMARY KEY,
    position INTEGER NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE TABLE base_references (
    id BIGSERIAL PRIMARY KEY,
    category_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);

CREATE TABLE book_references (
    id BIGINT PRIMARY KEY,
    reference_id BIGINT NOT NULL UNIQUE,
    isbn VARCHAR(50) NOT NULL,
    FOREIGN KEY (reference_id) REFERENCES base_references(id)
);

CREATE TABLE link_references (
    id BIGINT PRIMARY KEY,
    reference_id BIGINT NOT NULL UNIQUE,
    url TEXT NOT NULL,
    description TEXT,
    FOREIGN KEY (reference_id) REFERENCES base_references(id)
);

CREATE TABLE note_references (
    id BIGINT PRIMARY KEY,
    reference_id BIGINT NOT NULL UNIQUE,
    text TEXT NOT NULL,
    FOREIGN KEY (reference_id) REFERENCES base_references(id)
);

CREATE TABLE reference_positions (
    reference_id BIGINT PRIMARY KEY,
    category_id BIGINT NOT NULL,
    position INTEGER NOT NULL,
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
