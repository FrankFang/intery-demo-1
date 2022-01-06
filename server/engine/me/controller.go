package me

import (
	"intery/server/database"
	"intery/server/engine/base"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	base.BaseController
}

func (ctrl *Controller) Show(c *gin.Context) {
	userId, err := ctrl.MustSignIn(c)
	if err != nil {
		return
	}
	u := database.GetQuery().User
	user, err := u.WithContext(c).Where(u.ID.Eq(userId)).First()
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(200, gin.H{
		"resource": user,
	})
}
