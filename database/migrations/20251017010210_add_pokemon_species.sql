-- +goose Up
-- +goose StatementBegin
CREATE TABLE pokemon_species (
  species_id   SMALLSERIAL PRIMARY KEY,
  name         TEXT UNIQUE NOT NULL,
  base_attack  SMALLINT NOT NULL,
  base_defense SMALLINT NOT NULL,
  base_speed   SMALLINT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE pokemon_species IF EXISTS;
-- +goose StatementEnd
