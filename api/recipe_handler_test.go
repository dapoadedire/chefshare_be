package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dapoadedire/chefshare_be/store"
	"github.com/gin-gonic/gin"
)

// MockRecipeStore implements store.RecipeStore for testing
type MockRecipeStore struct {
	recipes        map[int64]*store.Recipe
	completeRecipe *store.CompleteRecipe
	listResponse   *store.RecipeListResponse
	error          error
}

func (m *MockRecipeStore) GetCompleteRecipe(id int64) (*store.CompleteRecipe, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.completeRecipe, nil
}

func (m *MockRecipeStore) GetRecipes(options *store.RecipeListOptions) (*store.RecipeListResponse, error) {
	if m.error != nil {
		return nil, m.error
	}
	return m.listResponse, nil
}

func (m *MockRecipeStore) CreateRecipe(recipe *store.Recipe) error {
	if m.error != nil {
		return m.error
	}
	recipe.ID = 1
	return nil
}

func (m *MockRecipeStore) GetRecipeByID(id int64) (*store.Recipe, error) {
	if m.error != nil {
		return nil, m.error
	}
	if recipe, exists := m.recipes[id]; exists {
		return recipe, nil
	}
	return nil, nil
}

func (m *MockRecipeStore) GetRecipesByUserID(userID int64) ([]*store.Recipe, error) {
	return nil, m.error
}

func (m *MockRecipeStore) GetRecipesByUsername(username string) ([]*store.Recipe, error) {
	if m.error != nil {
		return nil, m.error
	}
	var recipes []*store.Recipe
	for _, recipe := range m.recipes {
		recipes = append(recipes, recipe)
	}
	return recipes, nil
}

func (m *MockRecipeStore) UpdateRecipe(recipe *store.Recipe) error {
	return m.error
}

func (m *MockRecipeStore) DeleteRecipe(id int64) error {
	return m.error
}

// Add stubs for all other required methods
func (m *MockRecipeStore) AddRecipePhoto(photo *store.RecipePhoto) error { return m.error }
func (m *MockRecipeStore) GetRecipePhotos(recipeID int64) ([]*store.RecipePhoto, error) { return nil, m.error }
func (m *MockRecipeStore) SetPrimaryPhoto(photoID int64, recipeID int64) error { return m.error }
func (m *MockRecipeStore) DeleteRecipePhoto(photoID int64) error { return m.error }
func (m *MockRecipeStore) AddRecipeIngredient(ingredient *store.RecipeIngredient) error { return m.error }
func (m *MockRecipeStore) GetRecipeIngredients(recipeID int64) ([]*store.RecipeIngredient, error) { return nil, m.error }
func (m *MockRecipeStore) UpdateRecipeIngredient(ingredient *store.RecipeIngredient) error { return m.error }
func (m *MockRecipeStore) DeleteRecipeIngredient(ingredientID int64) error { return m.error }
func (m *MockRecipeStore) AddRecipeStep(step *store.RecipeStep) error { return m.error }
func (m *MockRecipeStore) GetRecipeSteps(recipeID int64) ([]*store.RecipeStep, error) { return nil, m.error }
func (m *MockRecipeStore) UpdateRecipeStep(step *store.RecipeStep) error { return m.error }
func (m *MockRecipeStore) DeleteRecipeStep(stepID int64) error { return m.error }
func (m *MockRecipeStore) AddRecipeTag(recipeID int64, tagID int64) error { return m.error }
func (m *MockRecipeStore) RemoveRecipeTag(recipeID int64, tagID int64) error { return m.error }
func (m *MockRecipeStore) GetRecipeTags(recipeID int64) ([]*store.Tag, error) { return nil, m.error }
func (m *MockRecipeStore) GetAllCategories() ([]*store.Category, error) { return nil, m.error }
func (m *MockRecipeStore) GetAllTags() ([]*store.Tag, error) { return nil, m.error }
func (m *MockRecipeStore) CreateTag(name string) (*store.Tag, error) { return nil, m.error }
func (m *MockRecipeStore) CreateCategory(name string) (*store.Category, error) { return nil, m.error }
func (m *MockRecipeStore) AddRecipeReview(recipeID int64, userID int64, rating int, comment string) error { return m.error }
func (m *MockRecipeStore) GetRecipeReviews(recipeID int64) ([]*store.RecipeReview, error) { return nil, m.error }
func (m *MockRecipeStore) UpdateRecipeReview(review *store.RecipeReview) error { return m.error }
func (m *MockRecipeStore) DeleteRecipeReview(reviewID int64) error { return m.error }

// MockUserStore for testing
type MockUserStore struct {
	error error
}

func (m *MockUserStore) CreateUser(user *store.User) error { return m.error }
func (m *MockUserStore) CreateUserWithTransaction(user *store.User, tx *sql.Tx) error { return m.error }
func (m *MockUserStore) GetUserByEmail(email string) (*store.User, error) { return nil, m.error }
func (m *MockUserStore) GetUserByID(userID string) (*store.User, error) { return nil, m.error }
func (m *MockUserStore) UpdatePassword(userID string, newPassword string) error { return m.error }
func (m *MockUserStore) UpdateUser(userID string, updates map[string]interface{}) (*store.User, error) { return nil, m.error }
func (m *MockUserStore) UpdateLastLogin(userID string) error { return m.error }
func (m *MockUserStore) IsUsernameTaken(username string, excludeUserID string) (bool, error) { return false, m.error }
func (m *MockUserStore) DB() *sql.DB { return nil }

func TestRecipeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("GetRecipes", func(t *testing.T) {
		mockStore := &MockRecipeStore{
			listResponse: &store.RecipeListResponse{
				Recipes:    []*store.Recipe{},
				TotalCount: 0,
				Page:       1,
				Limit:      10,
				TotalPages: 0,
			},
		}
		mockUserStore := &MockUserStore{}
		handler := NewRecipeHandler(mockStore, mockUserStore)

		router := gin.New()
		router.GET("/recipes", handler.GetRecipes)

		req, _ := http.NewRequest("GET", "/recipes", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		var response store.RecipeListResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Failed to unmarshal response: %v", err)
		}
	})

	t.Run("GetRecipe", func(t *testing.T) {
		mockStore := &MockRecipeStore{
			completeRecipe: &store.CompleteRecipe{
				Recipe: &store.Recipe{
					ID:    1,
					Title: "Test Recipe",
				},
			},
		}
		mockUserStore := &MockUserStore{}
		handler := NewRecipeHandler(mockStore, mockUserStore)

		router := gin.New()
		router.GET("/recipes/:id", handler.GetRecipe)

		req, _ := http.NewRequest("GET", "/recipes/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("GetRecipe_InvalidID", func(t *testing.T) {
		mockStore := &MockRecipeStore{}
		mockUserStore := &MockUserStore{}
		handler := NewRecipeHandler(mockStore, mockUserStore)

		router := gin.New()
		router.GET("/recipes/:id", handler.GetRecipe)

		req, _ := http.NewRequest("GET", "/recipes/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("CreateRecipe_Unauthorized", func(t *testing.T) {
		mockStore := &MockRecipeStore{}
		mockUserStore := &MockUserStore{}
		handler := NewRecipeHandler(mockStore, mockUserStore)

		router := gin.New()
		router.POST("/recipes", handler.CreateRecipe)

		reqBody := CreateRecipeRequest{
			Title:           "Test Recipe",
			DifficultyLevel: "easy",
			Status:          "draft",
		}
		jsonBody, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", w.Code)
		}
	})

	t.Run("GetUserRecipes", func(t *testing.T) {
		mockStore := &MockRecipeStore{
			recipes: map[int64]*store.Recipe{
				1: {
					ID:    1,
					Title: "User Recipe",
				},
			},
		}
		mockUserStore := &MockUserStore{}
		handler := NewRecipeHandler(mockStore, mockUserStore)

		router := gin.New()
		router.GET("/users/:username/recipes", handler.GetUserRecipes)

		req, _ := http.NewRequest("GET", "/users/testuser/recipes", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})
}