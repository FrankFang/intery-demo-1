package test

import (
	"fmt"
	"intery/db"
	"os"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
)

func Setup(t *testing.T, id int) {
	os.Setenv("PRIVATE_KEY", "/root/repos/intery-demo-1/intery.rsa")
	os.Setenv("PUBLIC_KEY", "/root/repos/intery-demo-1/intery.rsa.pub")
	os.Setenv("DB_HOST", "psql1")
	os.Setenv("DB_USER", "intery")
	os.Setenv("DB_NAME", fmt.Sprintf("intery_test_%d", id))
	os.Setenv("DB_PASSWORD", "123456")
	os.Setenv("DB_PORT", "5432")
	gin.SetMode(gin.TestMode)
	db.Create()
	db.Migrate()
}

type autoInc struct {
	sync.Mutex
	id int
}

var ai autoInc

func GetId() (id int) {
	ai.Lock()
	defer ai.Unlock()

	ai.id += 1
	id = ai.id
	return
}
