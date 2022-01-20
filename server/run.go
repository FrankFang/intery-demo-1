package server

import (
	"intery/server/database"
	"intery/server/engine"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

func Run() error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://00fb9af85b3e45a8bbc14818c6e6f9c6@o1120743.ingest.sentry.io/6156745",
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)
	sentry.CaptureMessage("It works!")


	err = database.Init()
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
