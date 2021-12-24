package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/JAbduvohidov/jwt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v4"
	"golang.org/x/crypto/bcrypt"
	"io"
	"log"
	"net/http"
	"sed/db"
	"sed/models"
	"strconv"
	"strings"
	"time"
)

var Validate *validator.Validate

func Authorization(c *gin.Context) {
	response := models.Response{
		Code:    http.StatusUnauthorized,
		Message: http.StatusText(http.StatusUnauthorized),
		Time:    time.Now(),
	}

	header := c.GetHeader("Authorization")

	parts := strings.Split(header, " ")

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
	err := db.Pool.QueryRow(
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

func Login(c *gin.Context) {
	var (
		employee models.Employee
		response = models.Response{
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

	err = Validate.Struct(employee)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	id := 0
	passwordHash := ""
	role := ""
	err = db.Pool.QueryRow(
		context.Background(),
		`select e.id, e.password, rg.role
from employees e
         left join role_group rg on e.role_id = rg.id
where email = $1;`,
		employee.Email,
	).Scan(
		&id,
		&passwordHash,
		&role,
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

	token, err := jwt.Encode(models.Token{
		Id:    id,
		Email: employee.Email,
	}, jwt.Secret("secret"))
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	_, err = db.Pool.Exec(
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

	response.Payload = struct {
		Token string `json:"token"`
		Role  string `json:"role"`
	}{
		Token: token,
		Role:  role,
	}

	c.JSON(http.StatusOK, &response)
}

func GetUsers(c *gin.Context) {
	var (
		employees      []models.Employee
		employeeFilter = models.EmployeeFilter{}
		response       = models.Response{
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

	err = Validate.Struct(employeeFilter)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	rows, err := db.Pool.Query(
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
		employee := models.Employee{}
		err = rows.Scan(
			&employee.Id,
			&employee.FullName,
			&employee.Role.Role,
			&employee.Email,
			&employee.Department.Name,
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

func EditUser(c *gin.Context) {
	var (
		externalEmployee models.Employee
		internalEmployee models.Employee
		response         = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	userId := c.GetInt("user-id")

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.Println("unable to read body data:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}
	err = json.Unmarshal(data, &externalEmployee)
	if err != nil {
		log.Println("error unmarshaling externalEmployee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = Validate.Struct(externalEmployee)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	err = db.Pool.QueryRow(
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
	).Scan(
		&internalEmployee.Id,
		&internalEmployee.FullName,
		&internalEmployee.Role.Role,
		&internalEmployee.Email,
		&internalEmployee.Department.Name,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	if internalEmployee.Role.Role != "ADMIN" {
		response.Code = http.StatusBadRequest
		response.Message = "no access to this page"
		c.JSON(http.StatusOK, &response)
		return
	}

	rtn, err := db.Pool.Exec(
		c,
		`update employees
set full_name     = $1,
    role_id       = $2,
    department_id = $3
where id = $4;`,
		externalEmployee.FullName,
		externalEmployee.RoleId,
		externalEmployee.DepartmentId,
		userId,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	if rtn.RowsAffected() < 1 {
		response.Code = http.StatusBadRequest
		response.Message = pgx.ErrNoRows.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	c.JSON(http.StatusOK, &response)
}

func GetProfile(c *gin.Context) {
	var (
		employee = models.Employee{}
		response = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	userId := c.GetInt("user-id")
	err := db.Pool.QueryRow(
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
	).Scan(
		&employee.Id,
		&employee.FullName,
		&employee.Role.Role,
		&employee.Email,
		&employee.Department.Name,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	paramUserId, _ := strconv.Atoi(c.Param("id"))

	if employee.Id == paramUserId {
		response.Payload = employee
		c.JSON(http.StatusOK, &response)
		return
	}

	if employee.Role.Role != "ADMIN" {
		response.Code = http.StatusBadRequest
		response.Message = "no access to this page"
		c.JSON(http.StatusOK, &response)
		return
	}

	err = db.Pool.QueryRow(
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
		paramUserId,
	).Scan(
		&employee.Id,
		&employee.FullName,
		&employee.Role.Role,
		&employee.Email,
		&employee.Department.Name,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	response.Payload = employee

	c.JSON(http.StatusOK, &response)
}

func GetRoles(c *gin.Context) {
	var (
		roleGroups []models.RoleGroup
		response   = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	rows, err := db.Pool.Query(
		c,
		`select id, role
from role_group;`,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	for rows.Next() {
		roleGroup := models.RoleGroup{}
		err = rows.Scan(
			&roleGroup.Id,
			&roleGroup.Role,
		)

		if err != nil {
			response.Code = http.StatusInternalServerError
			response.Message = err.Error()
			c.JSON(http.StatusOK, &response)
			return
		}

		roleGroups = append(roleGroups, roleGroup)
	}

	response.Payload = roleGroups

	c.JSON(http.StatusOK, &response)
}
