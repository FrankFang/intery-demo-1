package middlewares

import (
	"intery/server/middlewares/cors"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func New() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		cors.New(),
		sentrygin.New(sentrygin.Options{}),
	}
}
