package main

import (
	"github.com/gin-gonic/gin"
	"time"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"time":    time.Now(),
		})
	})

	r.POST("/login", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "test",
		})
	})

	r.Run()
}
