package app

import (
	"database/sql"
	"log"

	"github.com/dapoadedire/chefshare_be/api"
	"github.com/dapoadedire/chefshare_be/migrations"
	"github.com/dapoadedire/chefshare_be/services"
	"github.com/dapoadedire/chefshare_be/store"
)

type Application struct {
	DB                 *sql.DB
	AuthHandler        *api.AuthHandler
	UserHandler        *api.UserHandler
	EmailService       *services.EmailService
	UserStore          store.UserStore
	PasswordResetStore store.PasswordResetStore
	RefreshTokenStore  store.RefreshTokenStore
	JWTService         *services.JWTService
}

func NewApplication() (*Application, error) {

	pgDB, err := store.Open()
	if err != nil {
		return nil, err
	}

	err = store.MigrateFS(pgDB, migrations.FS, ".")
	if err != nil {
		panic(err)
	}

	// Initialize email service
	emailService, err := services.NewEmailService()
	if err != nil {
		log.Printf("Warning: Email service could not be initialized: %v", err)
		// Continue without email service
	}

	userStore := store.NewPostgresUserStore(pgDB)
	passwordResetStore := store.NewPostgresPasswordResetStore(pgDB)
	refreshTokenStore := store.NewPostgresRefreshTokenStore(pgDB)
	emailVerificationStore := store.NewPostgresEmailVerificationStore(pgDB)
	
	// Initialize JWT service with default configuration
	jwtConfig := services.DefaultJWTConfig()
	jwtService := services.NewJWTService(jwtConfig, refreshTokenStore, userStore)

	// This will be fully removed in a future update
	authHandler := api.NewAuthHandler(
		userStore, 
		refreshTokenStore, 
		passwordResetStore, 
		emailVerificationStore,
		emailService, 
		jwtService,
	)
	userHandler := api.NewUserHandler(userStore, emailService, jwtService)

	app := &Application{
		DB:                 pgDB,
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		EmailService:       emailService,
		UserStore:          userStore,
		PasswordResetStore: passwordResetStore,
		RefreshTokenStore:  refreshTokenStore,
		JWTService:         jwtService,
	}

	return app, nil
}
