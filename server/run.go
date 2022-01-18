package server

import (
	"intery/server/database"
	"intery/server/engine"
	"os"
)

func Run() error {
	err := database.Init()
	if err != nil {
		panic(err)
	}
	app, err := engine.Init()
	if err != nil {
		return err
	} else {

		if socket := os.Getenv("SOCKET"); socket != "" {
			return app.RunUnix(socket)
		} else {
			return app.Run("0.0.0.0:8080")
		}
	}
}
