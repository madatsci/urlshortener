-- +goose Up
-- +goose StatementBegin
CREATE TABLE urls (
    id uuid PRIMARY KEY,
    correlation_id character varying(255) NOT NULL DEFAULT '',
    slug character varying(255) NOT NULL,
    original_url text NOT NULL,
    created_at timestamp without time zone NOT NULL
);

CREATE UNIQUE INDEX urls_slug ON urls (slug);
CREATE UNIQUE INDEX urls_original_url ON urls (original_url)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE urls;
-- +goose StatementEnd
