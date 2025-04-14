-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_urls (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    url_id uuid NOT NULL,
    is_deleted bool DEFAULT false,
    created_at timestamp without time zone NOT NULL
)
-- +goose StatementEnd
-- +goose StatementBegin
CREATE UNIQUE INDEX user_urls_user_id_url_id ON user_urls (user_id, url_id);
ALTER TABLE user_urls ADD CONSTRAINT user_id_constraint FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE user_urls ADD CONSTRAINT url_id_constraint FOREIGN KEY (url_id) REFERENCES urls(id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_urls;
-- +goose StatementEnd
