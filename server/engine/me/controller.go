package me

import "github.com/gin-gonic/gin"

type Controller struct {
	User string
}

func (ctrl *Controller) Show(c *gin.Context) {
	ctrl.User = "me"
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
