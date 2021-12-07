package engine

import (
	"github.com/gin-gonic/gin"
)

func Init() (*gin.Engine, error) {
	gin.SetMode("debug")
	r := NewRouter()
	err := r.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}
