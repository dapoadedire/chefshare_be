-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS recipe_photos (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    recipe_id BIGINT NOT NULL,
    photo_url VARCHAR(255) NOT NULL,
    is_primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_recipe_photos_recipes FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE
);

CREATE INDEX idx_recipe_photos_recipe_id ON recipe_photos(recipe_id);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    recipe_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    image VARCHAR(255),
    quantity DOUBLE PRECISION,
    unit VARCHAR(50),
    position INTEGER, 
    CONSTRAINT fk_recipe_ingredients_recipes FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE
);

CREATE INDEX idx_recipe_ingredients_recipe_id ON recipe_ingredients(recipe_id);

CREATE TABLE IF NOT EXISTS recipe_steps (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    recipe_id BIGINT NOT NULL,
    step_number INTEGER NOT NULL,
    instruction TEXT NOT NULL,
    duration_in_minutes INTEGER,
    CONSTRAINT fk_recipe_steps_recipes FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE
);

CREATE INDEX idx_recipe_steps_recipe_id ON recipe_steps(recipe_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipe_steps;
DROP TABLE IF EXISTS recipe_ingredients;
DROP TABLE IF EXISTS recipe_photos;
-- +goose StatementEnd
