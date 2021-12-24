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
	"strconv"
	"time"
)

func GetAgreements(c *gin.Context) {
	var (
		agreements []models.Agreement
		response   = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	userId := c.GetInt("user-id")

	rows, err := db.Pool.Query(
		c,
		`select a.id, l.id, d.id, l.name, l.entry_date, a.viewed, coalesce(a.agreed_at, now())
from agreements a
         left join letters l on a.letter_id = l.id
         left join departments d on a.department_id = d.id
         left join described_letters dl on d.id = dl.department_id and dl.letter_id = l.id
where dl.executive_employee = $1
order by a.id desc;`,
		userId,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	for rows.Next() {
		agreement := models.Agreement{}

		err = rows.Scan(
			&agreement.Id,
			&agreement.Letter.Id,
			&agreement.DepartmentId,
			&agreement.Letter.Name,
			&agreement.Letter.EntryDate,
			&agreement.Viewed,
			&agreement.AgreedAt,
		)
		if err != nil {
			response.Code = http.StatusInternalServerError
			response.Message = err.Error()
			c.JSON(http.StatusOK, &response)
			return
		}

		agreements = append(agreements, agreement)
	}

	response.Payload = agreements

	c.JSON(http.StatusOK, &response)
}

func CreateAgreement(c *gin.Context) {
	var (
		agreement models.Agreement
		response  = models.Response{
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
	err = json.Unmarshal(data, &agreement)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	rtn, err := db.Pool.Exec(
		c,
		`insert into agreements (department_id, letter_id, viewed, agreed_at)
values ($1, $2, false, now());`,
		agreement.DepartmentId,
		agreement.LetterId,
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

func GetAgreement(c *gin.Context) {
	var (
		agreement models.Agreement
		response  = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	id, _ := strconv.Atoi(c.Param("id"))

	rtn, err := db.Pool.Exec(
		c,
		`update agreements
set viewed = true
where id = $1;`,
		id,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	if rtn.RowsAffected() < 1 {
		response.Code = http.StatusInternalServerError
		response.Message = pgx.ErrNoRows.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	err = db.Pool.QueryRow(
		c,
		`select a.id,
       l.id,
       d.id,
       l.name,
       l.entry_date,
       a.viewed,
       a.agreed_at
from agreements a
         left join letters l on a.letter_id = l.id
         left join departments d on a.department_id = d.id
         left join described_letters dl on d.id = dl.department_id and dl.letter_id = l.id
where a.id = $1;`,
		id,
	).Scan(
		&agreement.Id,
		&agreement.Letter.Id,
		&agreement.DepartmentId,
		&agreement.Letter.Name,
		&agreement.Letter.EntryDate,
		&agreement.Viewed,
		&agreement.AgreedAt,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	response.Payload = agreement

	c.JSON(http.StatusOK, &response)
}
