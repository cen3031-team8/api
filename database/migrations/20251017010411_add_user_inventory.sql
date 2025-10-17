-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_inventory (
  user_id  BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  item     TEXT NOT NULL,
  quantity INTEGER NOT NULL CHECK (quantity >= 0),
  PRIMARY KEY (user_id, item)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_inventory IF EXISTS;
-- +goose StatementEnd
