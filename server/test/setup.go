package test

import (
	"intery/db"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func Setup(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db.Reset("intery")
	os.Setenv("PRIVATE_KEY", "/root/repos/intery-demo-1/intery.rsa")
	os.Setenv("PUBLIC_KEY", "/root/repos/intery-demo-1/intery.rsa.pub")
}
