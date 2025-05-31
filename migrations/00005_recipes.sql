-- +goose Up
-- +goose StatementBegin

-- Create enum for recipe status
DO $$ BEGIN
    CREATE TYPE recipe_status AS ENUM ('draft', 'published', 'archived');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Create enum for difficulty level
DO $$ BEGIN
    CREATE TYPE recipe_difficulty_level AS ENUM ('easy', 'medium', 'hard');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS recipes (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    user_id BIGINT NOT NULL,
    category_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    published_at TIMESTAMPTZ,
    status recipe_status NOT NULL DEFAULT 'draft',
    difficulty_level recipe_difficulty_level NOT NULL DEFAULT 'easy',
    serving_size INTEGER,
    prep_time INTEGER,
    cook_time INTEGER,
    total_time INTEGER,
    CONSTRAINT fk_recipes_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_recipes_categories FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

-- Indexes
CREATE INDEX idx_recipes_user_id ON recipes(user_id);
CREATE INDEX idx_recipes_category_id ON recipes(category_id);
CREATE INDEX idx_recipes_status ON recipes(status);
CREATE INDEX idx_recipes_difficulty_level ON recipes(difficulty_level);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS recipes;
DROP TYPE IF EXISTS recipe_status;
DROP TYPE IF EXISTS recipe_difficulty_level;
-- +goose StatementEnd
