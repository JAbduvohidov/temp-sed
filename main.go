package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()

	r.POST("/login", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "test",
		})
	})

	r.Run()
}
