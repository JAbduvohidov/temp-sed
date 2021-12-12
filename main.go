package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log"
	"sed/db"
	"sed/handlers"
)

func main() {
	handlers.Validate = validator.New()
	r := gin.Default()

	err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	r.GET("/ping", handlers.Ping)

	r.POST("/login", handlers.Login)

	r.GET("/letters", handlers.Authorization, handlers.GetDocuments)

	r.GET("/letter/{id}", handlers.Authorization, handlers.GetDocument)

	r.POST("/letter", handlers.Authorization, handlers.CreateDocument)

	r.GET("/letters/type", handlers.Authorization, handlers.GetLetterTypes)

	r.GET("/users", handlers.Authorization, handlers.GetUsers)

	r.GET("/users/me", handlers.Authorization, handlers.GetProfile)

	r.GET("/roles", handlers.Authorization, handlers.GetRoles)

	log.Fatalln(r.Run())
}
