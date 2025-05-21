package app

import (
	"database/sql"

	"github.com/dapoadedire/chefshare_be/api"
	"github.com/dapoadedire/chefshare_be/migrations"
	"github.com/dapoadedire/chefshare_be/store"
)

type Application struct {
	DB          *sql.DB
	UserHandler *api.UserHandler
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

	userStore := store.NewPostgresUserStore(pgDB)

	userHandler := api.NewUserHandler(userStore)

	app := &Application{
		DB: pgDB,
		UserHandler: userHandler,
		
	}

	return app, nil
}
