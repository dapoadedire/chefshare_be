package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type RecipeStatus string
type DifficultyLevel string

const (
	StatusDraft     RecipeStatus = "draft"
	StatusPublished RecipeStatus = "published"
	StatusArchived  RecipeStatus = "archived"

	DifficultyEasy   DifficultyLevel = "easy"
	DifficultyMedium DifficultyLevel = "medium"
	DifficultyHard   DifficultyLevel = "hard"
)

type Recipe struct {
	ID              int64           `json:"id"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	UserID          int64           `json:"user_id"`
	CategoryID      *int64          `json:"category_id,omitempty"`
	CategoryName    *string         `json:"category_name,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
	PublishedAt     *time.Time      `json:"published_at,omitempty"`
	Status          RecipeStatus    `json:"status"`
	DifficultyLevel DifficultyLevel `json:"difficulty_level"`
	ServingSize     *int            `json:"serving_size,omitempty"`
	PrepTime        *int            `json:"prep_time,omitempty"`
	CookTime        *int            `json:"cook_time,omitempty"`
	TotalTime       *int            `json:"total_time,omitempty"`
}

type RecipePhoto struct {
	ID        int64     `json:"id"`
	RecipeID  int64     `json:"recipe_id"`
	PhotoURL  string    `json:"photo_url"`
	IsPrimary bool      `json:"is_primary"`
	CreatedAt time.Time `json:"created_at"`
}

type RecipeIngredient struct {
	ID       int64    `json:"id"`
	RecipeID int64    `json:"recipe_id"`
	Name     string   `json:"name"`
	Image    *string  `json:"image,omitempty"`
	Quantity *float64 `json:"quantity,omitempty"`
	Unit     *string  `json:"unit,omitempty"`
	Position *int     `json:"position,omitempty"`
}

type RecipeStep struct {
	ID                int64  `json:"id"`
	RecipeID          int64  `json:"recipe_id"`
	StepNumber        int    `json:"step_number"`
	Instruction       string `json:"instruction"`
	DurationInMinutes *int   `json:"duration_in_minutes,omitempty"`
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Tag struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type RecipeReview struct {
	ID        int64     `json:"id"`
	RecipeID  int64     `json:"recipe_id"`
	UserID    int64     `json:"user_id"`
	Rating    int       `json:"rating"`
	Comment   string    `json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type CompleteRecipe struct {
	Recipe      *Recipe             `json:"recipe"`
	Ingredients []*RecipeIngredient `json:"ingredients"`
	Steps       []*RecipeStep       `json:"steps"`
	Photos      []*RecipePhoto      `json:"photos"`
	Tags        []*Tag              `json:"tags"`
	Reviews     []*RecipeReview     `json:"reviews"`
}

type RecipeListOptions struct {
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	Category     string `json:"category,omitempty"`
	Difficulty   string `json:"difficulty,omitempty"`
	SortBy       string `json:"sort_by,omitempty"` // created_at, updated_at, title
	SortOrder    string `json:"sort_order,omitempty"` // asc, desc
	Status       string `json:"status,omitempty"`
	UserID       *int64 `json:"user_id,omitempty"`
	Username     string `json:"username,omitempty"`
}

type RecipeListResponse struct {
	Recipes    []*Recipe `json:"recipes"`
	TotalCount int64     `json:"total_count"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	TotalPages int       `json:"total_pages"`
}

type RecipeStore interface {
	GetCompleteRecipe(id int64) (*CompleteRecipe, error)
	GetRecipes(options *RecipeListOptions) (*RecipeListResponse, error)

	CreateRecipe(recipe *Recipe) error
	GetRecipeByID(id int64) (*Recipe, error)
	GetRecipesByUserID(userID int64) ([]*Recipe, error)
	GetRecipesByUsername(username string) ([]*Recipe, error)
	UpdateRecipe(recipe *Recipe) error
	DeleteRecipe(id int64) error

	AddRecipePhoto(photo *RecipePhoto) error
	GetRecipePhotos(recipeID int64) ([]*RecipePhoto, error)
	SetPrimaryPhoto(photoID int64, recipeID int64) error
	DeleteRecipePhoto(photoID int64) error

	AddRecipeIngredient(ingredient *RecipeIngredient) error
	GetRecipeIngredients(recipeID int64) ([]*RecipeIngredient, error)
	UpdateRecipeIngredient(ingredient *RecipeIngredient) error
	DeleteRecipeIngredient(ingredientID int64) error

	AddRecipeStep(step *RecipeStep) error
	GetRecipeSteps(recipeID int64) ([]*RecipeStep, error)
	UpdateRecipeStep(step *RecipeStep) error
	DeleteRecipeStep(stepID int64) error

	AddRecipeTag(recipeID int64, tagID int64) error
	RemoveRecipeTag(recipeID int64, tagID int64) error
	GetRecipeTags(recipeID int64) ([]*Tag, error)

	GetAllCategories() ([]*Category, error)
	GetAllTags() ([]*Tag, error)
	CreateTag(name string) (*Tag, error)
	CreateCategory(name string) (*Category, error)

	AddRecipeReview(recipeID int64, userID int64, rating int, comment string) error
	GetRecipeReviews(recipeID int64) ([]*RecipeReview, error)
	UpdateRecipeReview(review *RecipeReview) error
	DeleteRecipeReview(reviewID int64) error
}



type PostgresRecipeStore struct {
	db *sql.DB
}

func NewPostgresRecipeStore(db *sql.DB) *PostgresRecipeStore {
	return &PostgresRecipeStore{
		db: db,
	}
}

func (s *PostgresRecipeStore) GetCompleteRecipe(id int64) (*CompleteRecipe, error) {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Safe to call if tx is already committed

	// Get the main recipe with category name
	recipeQuery := `
        SELECT 
            r.id, r.title, r.description, r.user_id, r.category_id,
            r.created_at, r.updated_at, r.published_at, r.status, 
            r.difficulty_level, r.serving_size, r.prep_time, r.cook_time, r.total_time,
            c.name as category_name
        FROM recipes r
        LEFT JOIN categories c ON r.category_id = c.id
        WHERE r.id = $1
    `

	recipe := &Recipe{}
	err = tx.QueryRow(recipeQuery, id).Scan(
		&recipe.ID,
		&recipe.Title,
		&recipe.Description,
		&recipe.UserID,
		&recipe.CategoryID,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
		&recipe.PublishedAt,
		&recipe.Status,
		&recipe.DifficultyLevel,
		&recipe.ServingSize,
		&recipe.PrepTime,
		&recipe.CookTime,
		&recipe.TotalTime,
		&recipe.CategoryName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	// Get all other components
	ingredients, err := s.GetRecipeIngredientsTx(tx, id)
	if err != nil {
		return nil, err
	}

	steps, err := s.GetRecipeStepsTx(tx, id)
	if err != nil {
		return nil, err
	}

	photos, err := s.GetRecipePhotosTx(tx, id)
	if err != nil {
		return nil, err
	}

	tags, err := s.GetRecipeTagsTx(tx, id)
	if err != nil {
		return nil, err
	}

	reviews, err := s.GetRecipeReviewsTx(tx, id)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &CompleteRecipe{
		Recipe:      recipe,
		Ingredients: ingredients,
		Steps:       steps,
		Photos:      photos,
		Tags:        tags,
		Reviews:     reviews,
	}, nil
}

func (s *PostgresRecipeStore) CreateRecipe(recipe *Recipe) error {
	query := `
        INSERT INTO recipes(
            title, description, user_id, category_id, 
            status, difficulty_level, serving_size, 
            prep_time, cook_time, total_time
        ) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id, created_at, updated_at
    `

	err := s.db.QueryRow(
		query,
		recipe.Title,
		recipe.Description,
		recipe.UserID,
		recipe.CategoryID,
		recipe.Status,
		recipe.DifficultyLevel,
		recipe.ServingSize,
		recipe.PrepTime,
		recipe.CookTime,
		recipe.TotalTime,
	).Scan(
		&recipe.ID,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create recipe: %w", err)
	}

	return nil
}

func (s *PostgresRecipeStore) GetRecipeByID(id int64) (*Recipe, error) {
	query := `
		SELECT 
			r.id, r.title, r.description, r.user_id, r.category_id,
			r.created_at, r.updated_at, r.published_at, r.status, 
			r.difficulty_level, r.serving_size, r.prep_time, r.cook_time, r.total_time,
			c.name as category_name
		FROM recipes r
		LEFT JOIN categories c ON r.category_id = c.id
		WHERE r.id = $1
	`
	recipe := &Recipe{}
	err := s.db.QueryRow(query, id).Scan(
		&recipe.ID,
		&recipe.Title,
		&recipe.Description,
		&recipe.UserID,
		&recipe.CategoryID,
		&recipe.CategoryName,
		&recipe.CreatedAt,
		&recipe.UpdatedAt,
		&recipe.PublishedAt,
		&recipe.Status,
		&recipe.DifficultyLevel,
		&recipe.ServingSize,
		&recipe.PrepTime,
		&recipe.CookTime,
		&recipe.TotalTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	return recipe, nil
}

func (s *PostgresRecipeStore) GetRecipesByUserID(userID int64) ([]*Recipe, error) {
	query := `
		SELECT 
			r.id, r.title, r.description, r.user_id, r.category_id,
			r.created_at, r.updated_at, r.published_at, r.status, 
			r.difficulty_level, r.serving_size, r.prep_time, r.cook_time, r.total_time,
			c.name as category_name
		FROM recipes r
		LEFT JOIN categories c ON r.category_id = c.id
		WHERE r.user_id = $1
	`

	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes by user ID: %w", err)
	}
	defer rows.Close()

	var recipes []*Recipe
	for rows.Next() {
		recipe := &Recipe{}
		err := rows.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Description,
			&recipe.UserID,
			&recipe.CategoryID,
			&recipe.CategoryName,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
			&recipe.PublishedAt,
			&recipe.Status,
			&recipe.DifficultyLevel,
			&recipe.ServingSize,
			&recipe.PrepTime,
			&recipe.CookTime,
			&recipe.TotalTime,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %w", err)
		}

		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipes: %w", err)
	}

	return recipes, nil
}

func (s *PostgresRecipeStore) GetRecipes(options *RecipeListOptions) (*RecipeListResponse, error) {
	// Set defaults
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.Limit <= 0 || options.Limit > 50 {
		options.Limit = 10
	}
	if options.SortBy == "" {
		options.SortBy = "created_at"
	}
	if options.SortOrder == "" {
		options.SortOrder = "desc"
	}

	// Build WHERE clause
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	if options.Category != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("c.name = $%d", argIndex))
		args = append(args, options.Category)
		argIndex++
	}

	if options.Difficulty != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("r.difficulty_level = $%d", argIndex))
		args = append(args, options.Difficulty)
		argIndex++
	}

	if options.Status != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("r.status = $%d", argIndex))
		args = append(args, options.Status)
		argIndex++
	} else {
		// Default to published recipes only for public listing
		whereConditions = append(whereConditions, fmt.Sprintf("r.status = $%d", argIndex))
		args = append(args, "published")
		argIndex++
	}

	if options.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("r.user_id = $%d", argIndex))
		args = append(args, *options.UserID)
		argIndex++
	}

	if options.Username != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("u.username = $%d", argIndex))
		args = append(args, options.Username)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Validate sort fields
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
	}
	if !validSortFields[options.SortBy] {
		options.SortBy = "created_at"
	}

	validSortOrders := map[string]bool{
		"asc":  true,
		"desc": true,
	}
	if !validSortOrders[options.SortOrder] {
		options.SortOrder = "desc"
	}

	// Count total recipes
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM recipes r
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN users u ON r.user_id = u.id
		%s
	`, whereClause)

	var totalCount int64
	err := s.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count recipes: %w", err)
	}

	// Calculate pagination
	offset := (options.Page - 1) * options.Limit
	totalPages := int((totalCount + int64(options.Limit) - 1) / int64(options.Limit))

	// Get recipes
	query := fmt.Sprintf(`
		SELECT 
			r.id, r.title, r.description, r.user_id, r.category_id,
			r.created_at, r.updated_at, r.published_at, r.status, 
			r.difficulty_level, r.serving_size, r.prep_time, r.cook_time, r.total_time,
			c.name as category_name
		FROM recipes r
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN users u ON r.user_id = u.id
		%s
		ORDER BY r.%s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, options.SortBy, options.SortOrder, argIndex, argIndex+1)

	args = append(args, options.Limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes: %w", err)
	}
	defer rows.Close()

	var recipes []*Recipe
	for rows.Next() {
		recipe := &Recipe{}
		err := rows.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Description,
			&recipe.UserID,
			&recipe.CategoryID,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
			&recipe.PublishedAt,
			&recipe.Status,
			&recipe.DifficultyLevel,
			&recipe.ServingSize,
			&recipe.PrepTime,
			&recipe.CookTime,
			&recipe.TotalTime,
			&recipe.CategoryName,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %w", err)
		}

		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipes: %w", err)
	}

	return &RecipeListResponse{
		Recipes:    recipes,
		TotalCount: totalCount,
		Page:       options.Page,
		Limit:      options.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *PostgresRecipeStore) GetRecipesByUsername(username string) ([]*Recipe, error) {
	query := `
		SELECT 
			r.id, r.title, r.description, r.user_id, r.category_id,
			r.created_at, r.updated_at, r.published_at, r.status, 
			r.difficulty_level, r.serving_size, r.prep_time, r.cook_time, r.total_time,
			c.name as category_name
		FROM recipes r
		LEFT JOIN categories c ON r.category_id = c.id
		LEFT JOIN users u ON r.user_id = u.id
		WHERE u.username = $1 AND r.status = 'published'
		ORDER BY r.created_at DESC
	`

	rows, err := s.db.Query(query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipes by username: %w", err)
	}
	defer rows.Close()

	var recipes []*Recipe
	for rows.Next() {
		recipe := &Recipe{}
		err := rows.Scan(
			&recipe.ID,
			&recipe.Title,
			&recipe.Description,
			&recipe.UserID,
			&recipe.CategoryID,
			&recipe.CreatedAt,
			&recipe.UpdatedAt,
			&recipe.PublishedAt,
			&recipe.Status,
			&recipe.DifficultyLevel,
			&recipe.ServingSize,
			&recipe.PrepTime,
			&recipe.CookTime,
			&recipe.TotalTime,
			&recipe.CategoryName,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %w", err)
		}

		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipes: %w", err)
	}

	return recipes, nil
}

func (s *PostgresRecipeStore) UpdateRecipe(recipe *Recipe) error {
	query := `
		UPDATE recipes
		SET 
			title = $1, 
			description = $2, 
			category_id = $3, 
			status = $4, 
			difficulty_level = $5, 
			serving_size = $6, 
			prep_time = $7, 
			cook_time = $8, 
			total_time = $9,
			updated_at = NOW()
		WHERE id = $10
	`

	result, err := s.db.Exec(
		query,
		recipe.Title,
		recipe.Description,
		recipe.CategoryID,
		recipe.Status,
		recipe.DifficultyLevel,
		recipe.ServingSize,
		recipe.PrepTime,
		recipe.CookTime,
		recipe.TotalTime,
		recipe.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update recipe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) DeleteRecipe(id int64) error {
	query := `
		DELETE FROM recipes
		WHERE id = $1
	`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *PostgresRecipeStore) AddRecipePhoto(photo *RecipePhoto) error {
	query := `
		INSERT INTO recipe_photos (recipe_id, photo_url, is_primary)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := s.db.QueryRow(
		query,
		photo.RecipeID,
		photo.PhotoURL,
		photo.IsPrimary,
	).Scan(&photo.ID, &photo.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to add recipe photo: %w", err)
	}

	return nil
}

func (s *PostgresRecipeStore) GetRecipePhotos(recipeID int64) ([]*RecipePhoto, error) {
	query := `
		SELECT id, recipe_id, photo_url, is_primary, created_at
		FROM recipe_photos
		WHERE recipe_id = $1
	`

	rows, err := s.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe photos: %w", err)
	}
	defer rows.Close()

	var photos []*RecipePhoto
	for rows.Next() {
		photo := &RecipePhoto{}
		err := rows.Scan(&photo.ID, &photo.RecipeID, &photo.PhotoURL, &photo.IsPrimary, &photo.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe photo: %w", err)
		}
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe photos: %w", err)
	}

	return photos, nil
}
func (s *PostgresRecipeStore) SetPrimaryPhoto(photoID int64, recipeID int64) error {
	query := `
		UPDATE recipe_photos
		SET is_primary = TRUE
		WHERE id = $1 AND recipe_id = $2
	`

	result, err := s.db.Exec(query, photoID, recipeID)
	if err != nil {
		return fmt.Errorf("failed to set primary photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) DeleteRecipePhoto(photoID int64) error {
	query := `
		DELETE FROM recipe_photos
		WHERE id = $1
	`

	result, err := s.db.Exec(query, photoID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe photo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *PostgresRecipeStore) AddRecipeIngredient(ingredient *RecipeIngredient) error {
	query := `
		INSERT INTO recipe_ingredients (recipe_id, name, image, quantity, unit, position)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := s.db.QueryRow(
		query,
		ingredient.RecipeID,
		ingredient.Name,
		ingredient.Image,
		ingredient.Quantity,
		ingredient.Unit,
		ingredient.Position,
	).Scan(&ingredient.ID)

	if err != nil {
		return fmt.Errorf("failed to add recipe ingredient: %w", err)
	}

	return nil
}
func (s *PostgresRecipeStore) GetRecipeIngredients(recipeID int64) ([]*RecipeIngredient, error) {
	query := `
		SELECT id, recipe_id, name, image, quantity, unit, position
		FROM recipe_ingredients
		WHERE recipe_id = $1
		ORDER BY position
	`

	rows, err := s.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []*RecipeIngredient
	for rows.Next() {
		ingredient := &RecipeIngredient{}
		err := rows.Scan(&ingredient.ID, &ingredient.RecipeID, &ingredient.Name, &ingredient.Image, &ingredient.Quantity, &ingredient.Unit, &ingredient.Position)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe ingredient: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe ingredients: %w", err)
	}

	return ingredients, nil
}

func (s *PostgresRecipeStore) UpdateRecipeIngredient(ingredient *RecipeIngredient) error {
	query := `
		UPDATE recipe_ingredients
		SET 
			name = $1, 
			image = $2, 
			quantity = $3, 
			unit = $4, 
			position = $5
		WHERE id = $6 AND recipe_id = $7
	`

	result, err := s.db.Exec(
		query,
		ingredient.Name,
		ingredient.Image,
		ingredient.Quantity,
		ingredient.Unit,
		ingredient.Position,
		ingredient.ID,
		ingredient.RecipeID,
	)

	if err != nil {
		return fmt.Errorf("failed to update recipe ingredient: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) DeleteRecipeIngredient(ingredientID int64) error {
	query := `
		DELETE FROM recipe_ingredients
		WHERE id = $1
	`

	result, err := s.db.Exec(query, ingredientID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe ingredient: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) AddRecipeStep(step *RecipeStep) error {
	query := `
		INSERT INTO recipe_steps (recipe_id, step_number, instruction, duration_in_minutes)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := s.db.QueryRow(
		query,
		step.RecipeID,
		step.StepNumber,
		step.Instruction,
		step.DurationInMinutes,
	).Scan(&step.ID)

	if err != nil {
		return fmt.Errorf("failed to add recipe step: %w", err)
	}

	return nil
}
func (s *PostgresRecipeStore) GetRecipeSteps(recipeID int64) ([]*RecipeStep, error) {
	query := `
		SELECT id, recipe_id, step_number, instruction, duration_in_minutes
		FROM recipe_steps
		WHERE recipe_id = $1
		ORDER BY step_number
	`

	rows, err := s.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe steps: %w", err)
	}
	defer rows.Close()

	var steps []*RecipeStep
	for rows.Next() {
		step := &RecipeStep{}
		err := rows.Scan(&step.ID, &step.RecipeID, &step.StepNumber, &step.Instruction, &step.DurationInMinutes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe step: %w", err)
		}
		steps = append(steps, step)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe steps: %w", err)
	}

	return steps, nil
}
func (s *PostgresRecipeStore) UpdateRecipeStep(step *RecipeStep) error {
	query := `
		UPDATE recipe_steps
		SET 
			step_number = $1, 
			instruction = $2, 
			duration_in_minutes = $3
		WHERE id = $4 AND recipe_id = $5
	`

	result, err := s.db.Exec(
		query,
		step.StepNumber,
		step.Instruction,
		step.DurationInMinutes,
		step.ID,
		step.RecipeID,
	)

	if err != nil {
		return fmt.Errorf("failed to update recipe step: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) DeleteRecipeStep(stepID int64) error {
	query := `
		DELETE FROM recipe_steps
		WHERE id = $1
	`

	result, err := s.db.Exec(query, stepID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe step: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) AddRecipeTag(recipeID int64, tagID int64) error {
	query := `
		INSERT INTO recipe_tags (recipe_id, tag_id)
		VALUES ($1, $2)
	`

	_, err := s.db.Exec(query, recipeID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add recipe tag: %w", err)
	}

	return nil
}
func (s *PostgresRecipeStore) RemoveRecipeTag(recipeID int64, tagID int64) error {
	query := `
		DELETE FROM recipe_tags
		WHERE recipe_id = $1 AND tag_id = $2
	`

	result, err := s.db.Exec(query, recipeID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove recipe tag: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) GetRecipeTags(recipeID int64) ([]*Tag, error) {
	query := `
		SELECT t.id, t.name
		FROM recipe_tags rt
		JOIN tags t ON rt.tag_id = t.id
		WHERE rt.recipe_id = $1
	`

	rows, err := s.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe tags: %w", err)
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		tag := &Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe tags: %w", err)
	}

	return tags, nil
}
func (s *PostgresRecipeStore) GetAllCategories() ([]*Category, error) {
	query := `
		SELECT id, name
		FROM categories
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer rows.Close()

	var categories []*Category
	for rows.Next() {
		category := &Category{}
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over categories: %w", err)
	}
	return categories, nil
}
func (s *PostgresRecipeStore) GetAllTags() ([]*Tag, error) {
	query := `
		SELECT id, name
		FROM tags
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		tag := &Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over tags: %w", err)
	}
	return tags, nil
}
func (s *PostgresRecipeStore) CreateTag(name string) (*Tag, error) {
	query := `
		INSERT INTO tags (name)
		VALUES ($1)
		RETURNING id
	`

	tag := &Tag{Name: name}
	err := s.db.QueryRow(query, name).Scan(&tag.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}
func (s *PostgresRecipeStore) CreateCategory(name string) (*Category, error) {
	query := `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id
	`

	category := &Category{Name: name}
	err := s.db.QueryRow(query, name).Scan(&category.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}
func (s *PostgresRecipeStore) AddRecipeReview(recipeID int64, userID int64, rating int, comment string) error {
	query := `
		INSERT INTO recipe_reviews (recipe_id, user_id, rating, comment)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	review := &RecipeReview{
		RecipeID: recipeID,
		UserID:   userID,
		Rating:   rating,
		Comment:  comment,
	}

	err := s.db.QueryRow(
		query,
		review.RecipeID,
		review.UserID,
		review.Rating,
		review.Comment,
	).Scan(&review.ID, &review.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to add recipe review: %w", err)
	}

	return nil
}
func (s *PostgresRecipeStore) GetRecipeReviews(recipeID int64) ([]*RecipeReview, error) {
	query := `
		SELECT id, recipe_id, user_id, rating, comment, created_at
		FROM recipe_reviews
		WHERE recipe_id = $1
	`

	rows, err := s.db.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*RecipeReview
	for rows.Next() {
		review := &RecipeReview{}
		err := rows.Scan(&review.ID, &review.RecipeID, &review.UserID, &review.Rating, &review.Comment, &review.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe review: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe reviews: %w", err)
	}

	return reviews, nil
}
func (s *PostgresRecipeStore) UpdateRecipeReview(review *RecipeReview) error {
	query := `
		UPDATE recipe_reviews
		SET 
			rating = $1, 
			comment = $2, 
			created_at = NOW()
		WHERE id = $3 AND recipe_id = $4 AND user_id = $5
	`

	result, err := s.db.Exec(
		query,
		review.Rating,
		review.Comment,
		review.ID,
		review.RecipeID,
		review.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update recipe review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) DeleteRecipeReview(reviewID int64) error {
	query := `
		DELETE FROM recipe_reviews
		WHERE id = $1
	`

	result, err := s.db.Exec(query, reviewID)
	if err != nil {
		return fmt.Errorf("failed to delete recipe review: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
func (s *PostgresRecipeStore) GetRecipeIngredientsTx(tx *sql.Tx, recipeID int64) ([]*RecipeIngredient, error) {
	query := `
		SELECT id, recipe_id, name, image, quantity, unit, position
		FROM recipe_ingredients
		WHERE recipe_id = $1
		ORDER BY position
	`

	rows, err := tx.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe ingredients: %w", err)
	}
	defer rows.Close()

	var ingredients []*RecipeIngredient
	for rows.Next() {
		ingredient := &RecipeIngredient{}
		err := rows.Scan(
			&ingredient.ID,
			&ingredient.RecipeID,
			&ingredient.Name,
			&ingredient.Image,
			&ingredient.Quantity,
			&ingredient.Unit,
			&ingredient.Position,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe ingredient: %w", err)
		}
		ingredients = append(ingredients, ingredient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe ingredients: %w", err)
	}

	return ingredients, nil
}
func (s *PostgresRecipeStore) GetRecipeStepsTx(tx *sql.Tx, recipeID int64) ([]*RecipeStep, error) {
	query := `
		SELECT id, recipe_id, step_number, instruction, duration_in_minutes
		FROM recipe_steps
		WHERE recipe_id = $1
		ORDER BY step_number
	`

	rows, err := tx.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe steps: %w", err)
	}
	defer rows.Close()

	var steps []*RecipeStep
	for rows.Next() {
		step := &RecipeStep{}
		err := rows.Scan(
			&step.ID,
			&step.RecipeID,
			&step.StepNumber,
			&step.Instruction,
			&step.DurationInMinutes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe step: %w", err)
		}
		steps = append(steps, step)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe steps: %w", err)
	}

	return steps, nil
}
func (s *PostgresRecipeStore) GetRecipePhotosTx(tx *sql.Tx, recipeID int64) ([]*RecipePhoto, error) {
	query := `
		SELECT id, recipe_id, photo_url, is_primary, created_at
		FROM recipe_photos
		WHERE recipe_id = $1
	`

	rows, err := tx.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe photos: %w", err)
	}
	defer rows.Close()

	var photos []*RecipePhoto
	for rows.Next() {
		photo := &RecipePhoto{}
		err := rows.Scan(
			&photo.ID,
			&photo.RecipeID,
			&photo.PhotoURL,
			&photo.IsPrimary,
			&photo.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe photo: %w", err)
		}
		photos = append(photos, photo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe photos: %w", err)
	}

	return photos, nil
}
func (s *PostgresRecipeStore) GetRecipeTagsTx(tx *sql.Tx, recipeID int64) ([]*Tag, error) {
	query := `
		SELECT t.id, t.name
		FROM recipe_tags rt
		JOIN tags t ON rt.tag_id = t.id
		WHERE rt.recipe_id = $1
	`

	rows, err := tx.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe tags: %w", err)
	}
	defer rows.Close()

	var tags []*Tag
	for rows.Next() {
		tag := &Tag{}
		err := rows.Scan(&tag.ID, &tag.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe tags: %w", err)
	}

	return tags, nil
}
func (s *PostgresRecipeStore) GetRecipeReviewsTx(tx *sql.Tx, recipeID int64) ([]*RecipeReview, error) {
	query := `
		SELECT id, recipe_id, user_id, rating, comment, created_at
		FROM recipe_reviews
		WHERE recipe_id = $1
	`

	rows, err := tx.Query(query, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe reviews: %w", err)
	}
	defer rows.Close()

	var reviews []*RecipeReview
	for rows.Next() {
		review := &RecipeReview{}
		err := rows.Scan(
			&review.ID,
			&review.RecipeID,
			&review.UserID,
			&review.Rating,
			&review.Comment,
			&review.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recipe review: %w", err)
		}
		reviews = append(reviews, review)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipe reviews: %w", err)
	}

	return reviews, nil
}
