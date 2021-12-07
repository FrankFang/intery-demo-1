package server

import (
	"intery/server/config"
	"intery/server/database"
	"intery/server/engine"
)

func Run() error {
	config.Init()
	err := database.Init()
	if err != nil {
		panic(err)
	}
	app, err := engine.Init()
	if err != nil {
		return err
	} else {
		return app.Run(config.GetString("port"))
	}
}
