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

	r.POST("/ping", handlers.Ping)

	r.POST("/login", handlers.Login)

	r.POST("/letters", handlers.Authorization, handlers.GetDocuments)

	r.POST("/letter/:id", handlers.Authorization, handlers.GetDocument)

	r.POST("/letter", handlers.Authorization, handlers.CreateDocument)

	r.PUT("/letter", handlers.Authorization, handlers.EditDocument)

	r.POST("/letters/types", handlers.Authorization, handlers.GetLetterTypes)

	r.POST("/users", handlers.Authorization, handlers.GetUsers)

	r.POST("/users/:id", handlers.Authorization, handlers.GetProfile)

	r.PUT("/user", handlers.Authorization, handlers.EditUser)

	r.POST("/roles", handlers.Authorization, handlers.GetRoles)

	r.POST("/departments", handlers.Authorization, handlers.GetDepartments)

	r.POST("/department", handlers.Authorization, handlers.CreateDepartment)

	r.PUT("/department", handlers.Authorization, handlers.EditDepartment)

	r.POST("/departments/:id/users", handlers.Authorization, handlers.GetDepartmentEmployees)

	r.POST("/letters/describe", handlers.Authorization, handlers.DescribeLetter)

	r.POST("/letters/agreements", handlers.Authorization, handlers.GetAgreements)

	r.POST("/letters/agreement", handlers.Authorization, handlers.CreateAgreement)

	r.POST("/letters/agreement/:id", handlers.Authorization, handlers.GetAgreement)

	log.Fatalln(r.Run())
}
