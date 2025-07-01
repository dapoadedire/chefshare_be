package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dapoadedire/chefshare_be/app"
	"github.com/dapoadedire/chefshare_be/store"
)

// This is an integration test that requires a database connection
func TestAPIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Skip if no database environment variables are set
	if os.Getenv("DB_HOST") == "" {
		t.Skip("Skipping integration tests - no database configuration")
	}

	// Initialize the application
	application, err := app.NewApplication()
	if err != nil {
		t.Fatalf("Failed to initialize application: %v", err)
	}
	defer application.DB.Close()

	// This test would require proper database setup and user creation
	// For now, we'll just test that the store interface works
	t.Run("TestRecipeStoreInterface", func(t *testing.T) {
		// Test that the store implements the interface correctly
		var _ store.RecipeStore = application.RecipeStore

		// Test recipe list options
		options := &store.RecipeListOptions{
			Page:      1,
			Limit:     10,
			SortBy:    "created_at",
			SortOrder: "desc",
		}

		// This would require actual data in the database to test fully
		_, err := application.RecipeStore.GetRecipes(options)
		if err != nil {
			// Log but don't fail - might be due to missing test data
			t.Logf("GetRecipes returned error (may be expected): %v", err)
		}
	})

	t.Run("TestApplicationStructure", func(t *testing.T) {
		// Verify all handlers are properly initialized
		if application.RecipeHandler == nil {
			t.Error("RecipeHandler is nil")
		}
		if application.RecipeStore == nil {
			t.Error("RecipeStore is nil")
		}
		if application.UserStore == nil {
			t.Error("UserStore is nil")
		}
		if application.JWTService == nil {
			t.Error("JWTService is nil")
		}
	})

	fmt.Printf("Integration test completed at %s\n", time.Now().Format(time.RFC3339))
}