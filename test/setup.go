package test

import (
	"fmt"
	"intery/db"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func Setup(t *testing.T, id string) {
	os.Setenv("PRIVATE_KEY", "/root/repos/intery-demo-1/intery.rsa")
	os.Setenv("PUBLIC_KEY", "/root/repos/intery-demo-1/intery.rsa.pub")
	os.Setenv("DB_HOST", "psql1")
	os.Setenv("DB_USER", "intery")
	os.Setenv("DB_NAME", fmt.Sprintf("test_%s", id))
	os.Setenv("DB_PASSWORD", "123456")
	os.Setenv("DB_PORT", "5432")
	gin.SetMode(gin.TestMode)
	db.Create()
	db.Migrate()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz_")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func GetId() string {
	rand.Seed(time.Now().UnixNano())
	return randSeq(20)
}
