package hi

import (
	"github.com/gin-gonic/gin"
)

type Controller struct{}

func (ctrl Controller) Show(c *gin.Context) {

	c.JSON(200, gin.H{
		"message": "Hello World",
	})
}
