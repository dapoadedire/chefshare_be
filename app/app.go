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
	DB           *sql.DB
	AuthHandler  *api.AuthHandler
	EmailService *services.EmailService
	SessionStore store.SessionStore
	UserStore    store.UserStore
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

	authHandler := api.NewAuthHandler(userStore, sessionStore, emailService)

	app := &Application{
		DB:           pgDB,
		AuthHandler:  authHandler,
		EmailService: emailService,
		UserStore:    userStore,
		SessionStore: sessionStore,
	}

	return app, nil
}
