-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id uuid PRIMARY KEY,
    created_at timestamp without time zone NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
-- +goose StatementEnd
