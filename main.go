package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JAbduvohidov/jwt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type DocumentType struct {
	Id   int    `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

type RoleGroup struct {
	Id   int    `json:"id,omitempty"`
	Role string `json:"role,omitempty"`
}

type Department struct {
	Id             int    `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	InternalNumber string `json:"internal_number,omitempty"`
	Phone          string `json:"phone,omitempty"`
}

type Employee struct {
	Id           int        `json:"id,omitempty"`
	FullName     string     `json:"full_name,omitempty"`
	RoleId       int        `json:"role_id,omitempty"`
	Role         RoleGroup  `json:"role,omitempty"`
	Email        string     `json:"email,omitempty" validate:"required,email"`
	Password     string     `json:"password,omitempty" validate:"required,min=8"`
	Token        string     `json:"token,omitempty"`
	DepartmentId int        `json:"department_id,omitempty"`
	Department   Department `json:"department,omitempty"`
}

type EmployeeFilter struct {
	FullName   string `json:"full_name" validate:"min=3"`
	Email      string `json:"email" validate:"email"`
	Department string `json:"department" validate:"min=1"`
	RowsLimit  uint   `json:"rows_limit" validate:"required,number,min=1"`
	RowsOffset uint   `json:"rows_offset" validate:"number,min=0"`
}

type Letter struct {
	Id                 int          `json:"id,omitempty"`
	Name               string       `json:"name,omitempty" validate:"required,min=2"`
	Sender             string       `json:"sender,omitempty" validate:"required,min=2"`
	DocumentTypeId     int          `json:"document_type_id,omitempty" validate:"required,number"`
	DocumentType       DocumentType `json:"document_type,omitempty"`
	RegistrationNumber string       `json:"registration_number,omitempty"`
	EntryDate          time.Time    `json:"entry_date,omitempty"`
	OutgoingNumber     string       `json:"outgoing_number,omitempty"`
	DistributionDate   time.Time    `json:"distribution_date,omitempty"`
	Content            string       `json:"content,omitempty" validate:"required,min=20"`
}

type DescribedLetter struct {
	Id                int        `json:"id,omitempty"`
	LetterId          int        `json:"letter_id,omitempty"`
	Letter            Letter     `json:"letter,omitempty"`
	DepartmentId      int        `json:"department_id,omitempty"`
	Department        Department `json:"department,omitempty"`
	ExecutiveEmployee int        `json:"executive_employee,omitempty"`
	Employee          Employee   `json:"employee,omitempty"`
}

type Agreement struct {
	Id           int        `json:"id,omitempty"`
	DepartmentId int        `json:"department_id,omitempty"`
	Department   Department `json:"department,omitempty"`
	LetterId     int        `json:"letter_id,omitempty"`
	Letter       Letter     `json:"letter,omitempty"`
	Viewed       bool       `json:"viewed,omitempty"`
	AgreedAt     time.Time  `json:"agreed_at,omitempty"`
}

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Time    time.Time   `json:"time"`
	Payload interface{} `json:"payload,omitempty"`
}

type Token struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

var validate *validator.Validate
var pool *pgxpool.Pool

func main() {
	var err error

	pool, err = pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		return
	}

	validate = validator.New()
	r := gin.Default()

	r.GET("/ping", ping)

	r.POST("/login", login)

	r.GET("/letters", authorization, getDocuments)

	r.GET("/letter/{id}", authorization, getDocument)

	r.POST("/letter", authorization, createDocument)

	r.GET("/letters/type", authorization, getLetterTypes)

	r.GET("/users", authorization, getUsers)

	r.GET("/users/me", authorization, getProfile)

	log.Fatalln(r.Run())
}

func ping(c *gin.Context) {
	response := Response{
		Code:    http.StatusOK,
		Message: "pong",
		Time:    time.Now(),
	}

	c.JSON(http.StatusOK, &response)
}

func login(c *gin.Context) {
	var (
		employee Employee
		response = Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("unable to read body data:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = json.Unmarshal(data, &employee)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = validate.Struct(employee)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	id := 0
	passwordHash := ""
	err = pool.QueryRow(
		context.Background(),
		`select id, password from employees where email = $1`,
		employee.Email,
	).Scan(
		&id,
		&passwordHash,
	)
	if err != nil {
		if errors.Is(pgx.ErrNoRows, err) {
			response.Code = http.StatusUnauthorized
			response.Message = "invalid user"
			c.JSON(http.StatusOK, &response)
			return
		}
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(employee.Password))
	if err != nil {
		response.Code = http.StatusUnauthorized
		response.Message = "invalid user password"
		c.JSON(http.StatusOK, &response)
		return
	}

	token, err := jwt.Encode(Token{
		Id:    id,
		Email: employee.Email,
	}, jwt.Secret("secret"))
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	_, err = pool.Exec(
		context.Background(),
		`update employees set token = $1 where email = $2;`,
		token,
		employee.Email,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	response.Payload = token

	c.JSON(http.StatusOK, &response)
}

func getDocuments(c *gin.Context) {
	var (
		documentLetters []Letter
		response        = Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	rows, err := pool.Query(
		c,
		`select l.id,
       l.name,
       l.sender,
       dt.type,
       l.registration_number,
       l.entry_date,
       l.outgoing_number,
       l.distribution_date
from document_letters l
         left join document_type dt on l.document_type_id = dt.id;`,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	for rows.Next() {
		letter := Letter{}
		err = rows.Scan(
			&letter.Id,
			&letter.Name,
			&letter.Sender,
			&letter.DocumentType,
			&letter.RegistrationNumber,
			&letter.EntryDate,
			&letter.OutgoingNumber,
			&letter.DistributionDate,
		)

		documentLetters = append(documentLetters, letter)
	}

	response.Payload = documentLetters

	c.JSON(http.StatusOK, &response)
}

func getDocument(c *gin.Context) {
	var (
		documentLetter Letter
		response       = Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	err := pool.QueryRow(
		c,
		`select l.id,
       l.name,
       l.sender,
       dt.type,
       l.registration_number,
       l.entry_date,
       l.outgoing_number,
       l.distribution_date,
       l.content
from letters l
         left join document_type dt on l.document_type_id = dt.id
where l.id = $1;`,
	).Scan(
		&documentLetter.Id,
		&documentLetter.Name,
		&documentLetter.Sender,
		&documentLetter.DocumentType,
		&documentLetter.RegistrationNumber,
		&documentLetter.EntryDate,
		&documentLetter.OutgoingNumber,
		&documentLetter.DistributionDate,
		&documentLetter.Content,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	response.Payload = documentLetter

	c.JSON(http.StatusOK, &response)
}

func createDocument(c *gin.Context) {
	var (
		documentLetter Letter
		response       = Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("unable to read body data:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = json.Unmarshal(data, &documentLetter)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = validate.Struct(documentLetter)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	rtn, err := pool.Exec(
		c,
		`insert into letters (name, sender, document_type_id, registration_number, entry_date, outgoing_number, content)
values ($1, $2, $3, now(), now(), now(), $4);`,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	if rtn.RowsAffected() < 1 {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	c.JSON(http.StatusOK, &response)
}

func getLetterTypes(c *gin.Context) {
	var (
		documentTypes []DocumentType
		response      = Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	rows, err := pool.Query(
		c,
		`select id, type
from document_type;`,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	for rows.Next() {
		documentType := DocumentType{}
		err = rows.Scan(
			&documentType.Id,
			&documentType.Type,
		)

		documentTypes = append(documentTypes, documentType)
	}

	response.Payload = documentTypes

	c.JSON(http.StatusOK, &response)
}

func getUsers(c *gin.Context) {
	var (
		employees      []Employee
		employeeFilter = EmployeeFilter{}
		response       = Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("unable to read body data:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = json.Unmarshal(data, &employeeFilter)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = validate.Struct(employeeFilter)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	//TODO: add filters
	rows, err := pool.Query(
		c,
		`select e.id, e.full_name, rg.role, e.email, d.name
	from employees e
	left join role_group rg on e.role_id = rg.id
	left join departments d on e.department_id = d.id
	order by e.id desc
	offset $1 limit $2; `,
		employeeFilter.RowsOffset,
		employeeFilter.RowsLimit,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	for rows.Next() {
		employee := Employee{}
		err = rows.Scan(
			&employee.Id,
			&employee.FullName,
			&employee.Role,
			&employee.Email,
			&employee.Department,
		)

		if err != nil {
			response.Code = http.StatusInternalServerError
			response.Message = err.Error()
			c.JSON(http.StatusOK, &response)
			return
		}

		employees = append(employees, employee)
	}

	response.Payload = employees

	c.JSON(http.StatusOK, &response)
}

func getProfile(c *gin.Context) {
	var (
		employee = Employee{}
		response = Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	userId := c.GetString("user-id")

	err := pool.QueryRow(
		c,
		`select e.id,
       e.full_name,
       rg.role,
       e.email,
       d.name
from employees e
         left join role_group rg on e.role_id = rg.id
         left join departments d on e.department_id = d.id
where e.id = $1;`,
		userId,
	).Scan()
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	response.Payload = employee

	c.JSON(http.StatusOK, &response)
}

func authorization(c *gin.Context) {
	response := Response{
		Code:    http.StatusUnauthorized,
		Message: http.StatusText(http.StatusUnauthorized),
		Time:    time.Now(),
	}

	header := c.GetHeader("Authorization")

	parts := strings.Split(header, "")

	if len(parts) != 2 {
		response.Message = "invalid authorization header"
		c.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	authorizationType, token := parts[0], parts[1]

	if authorizationType != "Bearer" {
		response.Message = "invalid authorization type"
		c.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if len(token) < 30 {
		response.Message = "invalid token"
		c.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	id := 0
	err := pool.QueryRow(
		c,
		`select id
from employees
where token = $1;`,
		token,
	).Scan(&id)
	if err != nil {
		response.Message = err.Error()
		c.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if id == 0 {
		response.Message = "user not found or invalid token"
		c.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	c.Set("user-id", id)

	c.Next()
}
