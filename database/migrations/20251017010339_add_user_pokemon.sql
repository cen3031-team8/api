-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_pokemon (
  up_id        BIGSERIAL PRIMARY KEY,
  user_id      BIGINT NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  species_id   SMALLINT NOT NULL REFERENCES pokemon_species(species_id),
  nickname     TEXT,
  level        SMALLINT NOT NULL CHECK (level BETWEEN 1 AND 100),
  xp           INTEGER NOT NULL CHECK (xp >= 0),
  attack_mod   SMALLINT NOT NULL,   -- individualized attack modifier
  defense_mod  SMALLINT NOT NULL,   -- individualized defense modifier
  speed_mod    SMALLINT NOT NULL   -- individualized speed modifier
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE user_pokemon IF EXISTS;
-- +goose StatementEnd
