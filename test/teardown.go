package test

import (
	"intery/db"
	"intery/server/database"
	"os"
	"testing"

	"gopkg.in/h2non/gock.v1"
)

func Teardown(t *testing.T, id string) {
	database.CloseDB()
	db.Drop()
	os.Unsetenv("DB_NAME")
	os.Unsetenv("PRIVATE_KEY")
	os.Unsetenv("PUBLIC_KEY")
	gock.Off()
}
