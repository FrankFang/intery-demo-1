package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode("release")
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "你好，世界",
		})
	})
	socket := os.Getenv("SOCKET")
	if socket == "" {
		log.Fatal("未指定 socket")
	}
	log.Println("I am listening on", socket)
	err := router.RunUnix(socket)
	if err != nil {
		log.Fatal(err)
	}
}
