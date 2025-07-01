package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/dapoadedire/chefshare_be/store"
	"github.com/gin-gonic/gin"
)

type CreateRecipeRequest struct {
	Title           string                        `json:"title" binding:"required" validate:"min=1,max=255"`
	Description     string                        `json:"description"`
	CategoryID      *int64                        `json:"category_id"`
	DifficultyLevel string                        `json:"difficulty_level" binding:"required" validate:"oneof=easy medium hard"`
	ServingSize     *int                          `json:"serving_size"`
	PrepTime        *int                          `json:"prep_time"`
	CookTime        *int                          `json:"cook_time"`
	TotalTime       *int                          `json:"total_time"`
	Status          string                        `json:"status" binding:"required" validate:"oneof=draft published"`
	Ingredients     []CreateRecipeIngredientRequest `json:"ingredients"`
	Steps           []CreateRecipeStepRequest       `json:"steps"`
	Photos          []CreateRecipePhotoRequest      `json:"photos"`
	TagIDs          []int64                       `json:"tag_ids"`
}

type CreateRecipeIngredientRequest struct {
	Name     string   `json:"name" binding:"required"`
	Image    *string  `json:"image"`
	Quantity *float64 `json:"quantity"`
	Unit     *string  `json:"unit"`
	Position *int     `json:"position"`
}

type CreateRecipeStepRequest struct {
	StepNumber        int    `json:"step_number" binding:"required"`
	Instruction       string `json:"instruction" binding:"required"`
	DurationInMinutes *int   `json:"duration_in_minutes"`
}

type CreateRecipePhotoRequest struct {
	PhotoURL  string `json:"photo_url" binding:"required"`
	IsPrimary bool   `json:"is_primary"`
}

type UpdateRecipeRequest struct {
	Title           *string `json:"title"`
	Description     *string `json:"description"`
	CategoryID      *int64  `json:"category_id"`
	DifficultyLevel *string `json:"difficulty_level"`
	ServingSize     *int    `json:"serving_size"`
	PrepTime        *int    `json:"prep_time"`
	CookTime        *int    `json:"cook_time"`
	TotalTime       *int    `json:"total_time"`
	Status          *string `json:"status"`
}

type RecipeHandler struct {
	RecipeStore store.RecipeStore
	UserStore   store.UserStore
}

func NewRecipeHandler(recipeStore store.RecipeStore, userStore store.UserStore) *RecipeHandler {
	return &RecipeHandler{
		RecipeStore: recipeStore,
		UserStore:   userStore,
	}
}

// GetRecipes godoc
// @Summary List recipes
// @Description Get a paginated list of recipes with filtering and sorting options
// @Tags Recipes
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of recipes per page" default(10)
// @Param category query string false "Filter by category name"
// @Param difficulty query string false "Filter by difficulty level" Enums(easy, medium, hard)
// @Param sort_by query string false "Sort field" Enums(created_at, updated_at, title) default(created_at)
// @Param sort_order query string false "Sort order" Enums(asc, desc) default(desc)
// @Success 200 {object} store.RecipeListResponse "List of recipes"
// @Failure 400 {object} map[string]string "Invalid request parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipes [get]
func (h *RecipeHandler) GetRecipes(c *gin.Context) {
	options := &store.RecipeListOptions{}
	
	// Parse query parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			options.Page = page
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			options.Limit = limit
		}
	}
	
	options.Category = c.Query("category")
	options.Difficulty = c.Query("difficulty")
	options.SortBy = c.Query("sort_by")
	options.SortOrder = c.Query("sort_order")
	
	response, err := h.RecipeStore.GetRecipes(options)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recipes"})
		return
	}
	
	c.JSON(http.StatusOK, response)
}

// GetRecipe godoc
// @Summary Get a recipe by ID
// @Description Get detailed information about a specific recipe
// @Tags Recipes
// @Accept json
// @Produce json
// @Param id path int true "Recipe ID"
// @Success 200 {object} store.CompleteRecipe "Recipe details"
// @Failure 400 {object} map[string]string "Invalid recipe ID"
// @Failure 404 {object} map[string]string "Recipe not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipes/{id} [get]
func (h *RecipeHandler) GetRecipe(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
		return
	}
	
	recipe, err := h.RecipeStore.GetCompleteRecipe(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recipe"})
		return
	}
	
	if recipe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	
	c.JSON(http.StatusOK, recipe)
}

// CreateRecipe godoc
// @Summary Create a new recipe
// @Description Create a new recipe with ingredients, steps, and photos
// @Tags Recipes
// @Accept json
// @Produce json
// @Param request body CreateRecipeRequest true "Recipe creation request"
// @Security BearerAuth
// @Success 201 {object} store.Recipe "Created recipe"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipes [post]
func (h *RecipeHandler) CreateRecipe(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := strconv.ParseInt(userIDValue.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var req CreateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate difficulty level
	validDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true}
	if !validDifficulties[req.DifficultyLevel] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid difficulty level"})
		return
	}
	
	// Validate status
	validStatuses := map[string]bool{"draft": true, "published": true}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}
	
	// Create the recipe
	recipe := &store.Recipe{
		Title:           strings.TrimSpace(req.Title),
		Description:     strings.TrimSpace(req.Description),
		UserID:          userID,
		CategoryID:      req.CategoryID,
		DifficultyLevel: store.DifficultyLevel(req.DifficultyLevel),
		ServingSize:     req.ServingSize,
		PrepTime:        req.PrepTime,
		CookTime:        req.CookTime,
		TotalTime:       req.TotalTime,
		Status:          store.RecipeStatus(req.Status),
	}
	
	err = h.RecipeStore.CreateRecipe(recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create recipe"})
		return
	}
	
	// Add ingredients
	for _, ing := range req.Ingredients {
		ingredient := &store.RecipeIngredient{
			RecipeID: recipe.ID,
			Name:     strings.TrimSpace(ing.Name),
			Image:    ing.Image,
			Quantity: ing.Quantity,
			Unit:     ing.Unit,
			Position: ing.Position,
		}
		if err := h.RecipeStore.AddRecipeIngredient(ingredient); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add ingredient"})
			return
		}
	}
	
	// Add steps
	for _, step := range req.Steps {
		recipeStep := &store.RecipeStep{
			RecipeID:          recipe.ID,
			StepNumber:        step.StepNumber,
			Instruction:       strings.TrimSpace(step.Instruction),
			DurationInMinutes: step.DurationInMinutes,
		}
		if err := h.RecipeStore.AddRecipeStep(recipeStep); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add step"})
			return
		}
	}
	
	// Add photos
	for _, photo := range req.Photos {
		recipePhoto := &store.RecipePhoto{
			RecipeID:  recipe.ID,
			PhotoURL:  strings.TrimSpace(photo.PhotoURL),
			IsPrimary: photo.IsPrimary,
		}
		if err := h.RecipeStore.AddRecipePhoto(recipePhoto); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add photo"})
			return
		}
	}
	
	// Add tags
	for _, tagID := range req.TagIDs {
		if err := h.RecipeStore.AddRecipeTag(recipe.ID, tagID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add tag"})
			return
		}
	}
	
	c.JSON(http.StatusCreated, recipe)
}

// UpdateRecipe godoc
// @Summary Update a recipe
// @Description Update an existing recipe (owner only)
// @Tags Recipes
// @Accept json
// @Produce json
// @Param id path int true "Recipe ID"
// @Param request body UpdateRecipeRequest true "Recipe update request"
// @Security BearerAuth
// @Success 200 {object} store.Recipe "Updated recipe"
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not recipe owner"
// @Failure 404 {object} map[string]string "Recipe not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipes/{id} [put]
func (h *RecipeHandler) UpdateRecipe(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := strconv.ParseInt(userIDValue.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	
	// Get recipe ID
	idParam := c.Param("id")
	recipeID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
		return
	}
	
	// Check if recipe exists and user owns it
	existingRecipe, err := h.RecipeStore.GetRecipeByID(recipeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recipe"})
		return
	}
	
	if existingRecipe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	
	if existingRecipe.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only update your own recipes"})
		return
	}
	
	var req UpdateRecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Update fields if provided
	if req.Title != nil {
		existingRecipe.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		existingRecipe.Description = strings.TrimSpace(*req.Description)
	}
	if req.CategoryID != nil {
		existingRecipe.CategoryID = req.CategoryID
	}
	if req.DifficultyLevel != nil {
		validDifficulties := map[string]bool{"easy": true, "medium": true, "hard": true}
		if !validDifficulties[*req.DifficultyLevel] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid difficulty level"})
			return
		}
		existingRecipe.DifficultyLevel = store.DifficultyLevel(*req.DifficultyLevel)
	}
	if req.ServingSize != nil {
		existingRecipe.ServingSize = req.ServingSize
	}
	if req.PrepTime != nil {
		existingRecipe.PrepTime = req.PrepTime
	}
	if req.CookTime != nil {
		existingRecipe.CookTime = req.CookTime
	}
	if req.TotalTime != nil {
		existingRecipe.TotalTime = req.TotalTime
	}
	if req.Status != nil {
		validStatuses := map[string]bool{"draft": true, "published": true}
		if !validStatuses[*req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
			return
		}
		existingRecipe.Status = store.RecipeStatus(*req.Status)
	}
	
	err = h.RecipeStore.UpdateRecipe(existingRecipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update recipe"})
		return
	}
	
	c.JSON(http.StatusOK, existingRecipe)
}

// DeleteRecipe godoc
// @Summary Delete a recipe
// @Description Delete an existing recipe (owner only)
// @Tags Recipes
// @Accept json
// @Produce json
// @Param id path int true "Recipe ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Recipe deleted successfully"
// @Failure 400 {object} map[string]string "Invalid recipe ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - not recipe owner"
// @Failure 404 {object} map[string]string "Recipe not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /recipes/{id} [delete]
func (h *RecipeHandler) DeleteRecipe(c *gin.Context) {
	// Get user ID from context
	userIDValue, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	
	userID, err := strconv.ParseInt(userIDValue.(string), 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}
	
	// Get recipe ID
	idParam := c.Param("id")
	recipeID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid recipe ID"})
		return
	}
	
	// Check if recipe exists and user owns it
	existingRecipe, err := h.RecipeStore.GetRecipeByID(recipeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recipe"})
		return
	}
	
	if existingRecipe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	
	if existingRecipe.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own recipes"})
		return
	}
	
	err = h.RecipeStore.DeleteRecipe(recipeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete recipe"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Recipe deleted successfully"})
}

// GetUserRecipes godoc
// @Summary Get recipes by username
// @Description Get all published recipes by a specific user
// @Tags Recipes
// @Accept json
// @Produce json
// @Param username path string true "Username"
// @Success 200 {array} store.Recipe "List of user's recipes"
// @Failure 400 {object} map[string]string "Invalid username"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{username}/recipes [get]
func (h *RecipeHandler) GetUserRecipes(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}
	
	recipes, err := h.RecipeStore.GetRecipesByUsername(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recipes"})
		return
	}
	
	c.JSON(http.StatusOK, recipes)
}