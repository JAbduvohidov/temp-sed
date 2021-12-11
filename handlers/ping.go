package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"sed/models"
	"time"
)

func Ping(c *gin.Context) {
	response := models.Response{
		Code:    http.StatusOK,
		Message: "pong",
		Time:    time.Now(),
	}

	c.JSON(http.StatusOK, &response)
}
