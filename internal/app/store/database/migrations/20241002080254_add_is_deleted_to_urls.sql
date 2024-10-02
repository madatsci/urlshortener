-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls ADD COLUMN is_deleted bool;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls DROP COLUMN is_deleted;
-- +goose StatementEnd
