package middlewares

import (
	"intery/server/middlewares/cors"

	"github.com/gin-gonic/gin"
)

func New() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(),
	}
}
