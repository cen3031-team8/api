-- +goose Up
-- +goose StatementBegin
CREATE TABLE moves (
  move_id  SMALLSERIAL PRIMARY KEY,
  name     TEXT UNIQUE NOT NULL,
  power    SMALLINT NOT NULL,
  accuracy SMALLINT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE moves IF EXISTS;
-- +goose StatementEnd
