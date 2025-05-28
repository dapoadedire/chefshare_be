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
	SessionStore       store.SessionStore
	UserStore          store.UserStore
	PasswordResetStore store.PasswordResetStore
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
	sessionStore := store.NewPostgresSessionStore(pgDB)
	passwordResetStore := store.NewPostgresPasswordResetStore(pgDB)

	authHandler := api.NewAuthHandler(userStore, sessionStore, passwordResetStore, emailService)
	userHandler := api.NewUserHandler(userStore, emailService)

	app := &Application{
		DB:                 pgDB,
		AuthHandler:        authHandler,
		UserHandler:        userHandler,
		EmailService:       emailService,
		UserStore:          userStore,
		SessionStore:       sessionStore,
		PasswordResetStore: passwordResetStore,
	}

	return app, nil
}
