package server

import (
	"intery/server/config"
	"intery/server/database"
	"intery/server/engine"
)

func Run() error {
	config.Init()
	db := database.Init()
	app, err := engine.Init(db)
	if err != nil {
		return err
	} else {
		return app.Run(config.GetString("port"))
	}
}
