package server

import (
	"intery/server/config"
	"intery/server/database"
	"intery/server/engine"
)

func Run() error {
	config.Init()
	database.Init()
	app, err := engine.Init()
	if err != nil {
		return err
	} else {
		return app.Run(config.GetString("port"))
	}
}
