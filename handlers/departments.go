package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4"
	"io"
	"log"
	"net/http"
	"sed/db"
	"sed/models"
	"time"
)

func GetDepartments(c *gin.Context) {
	var (
		departments []models.Department
		response    = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	rows, err := db.Pool.Query(
		c,
		`select id, name, internal_number, phone
from departments
order by id desc;`,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	for rows.Next() {
		department := models.Department{}

		err = rows.Scan(
			&department.Id,
			&department.Name,
			&department.InternalNumber,
			&department.Phone,
		)
		if err != nil {
			response.Code = http.StatusInternalServerError
			response.Message = err.Error()
			c.JSON(http.StatusOK, &response)
			return
		}

		departments = append(departments, department)
	}

	response.Payload = departments

	c.JSON(http.StatusOK, &response)
}

func CreateDepartment(c *gin.Context) {
	var (
		department models.Department
		response   = models.Response{
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
	err = json.Unmarshal(data, &department)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	rtn, err := db.Pool.Exec(
		c,
		`insert into departments (name, internal_number, phone)
values ($1, $2, $3);`,
		department.Name,
		department.InternalNumber,
		department.Phone,
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

func EditDepartment(c *gin.Context) {
	var (
		department models.Department
		response   = models.Response{
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
	err = json.Unmarshal(data, &department)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	rtn, err := db.Pool.Exec(
		c,
		`update departments
set name            = $1,
    internal_number = $2,
    phone           = $3
where id = $4;`,
		department.Name,
		department.InternalNumber,
		department.Phone,
		department.Id,
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

//func GetDepartmentEmployees(c *gin.Context) {
//	var (
//		employees []models.Employee
//		response  = models.Response{
//			Code:    http.StatusOK,
//			Message: http.StatusText(http.StatusOK),
//			Time:    time.Now(),
//		}
//	)
//
//	id := c.Param("id")
//
//	rtn, err := db.Pool.Exec(
//		c,
//		`update departments
//set name            = $1,
//    internal_number = $2,
//    phone           = $3
//where id = $4;`,
//		department.Name,
//		department.InternalNumber,
//		department.Phone,
//		department.Id,
//	)
//	if err != nil {
//		response.Code = http.StatusInternalServerError
//		response.Message = err.Error()
//		c.JSON(http.StatusOK, &response)
//		return
//	}
//
//	if rtn.RowsAffected() < 1 {
//		response.Code = http.StatusBadRequest
//		response.Message = pgx.ErrNoRows.Error()
//		c.JSON(http.StatusOK, &response)
//		return
//	}
//
//	c.JSON(http.StatusOK, &response)
//}
