package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"sed/db"
	"sed/models"
	"time"
)

func GetDocuments(c *gin.Context) {
	var (
		documentLetters []models.Letter
		documentFilter  models.LetterFilter
		response        = models.Response{
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

	err = json.Unmarshal(data, &documentFilter)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = Validate.Struct(documentFilter)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	//TODO: add filters
	rows, err := db.Pool.Query(
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
         left join document_type dt on l.document_type_id = dt.id
order by l.id desc
offset $1 limit $2;`,
	)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	for rows.Next() {
		letter := models.Letter{}
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

func GetDocument(c *gin.Context) {
	var (
		documentLetter models.Letter
		response       = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	err := db.Pool.QueryRow(
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

func CreateDocument(c *gin.Context) {
	var (
		documentLetter models.Letter
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

	err = json.Unmarshal(data, &documentLetter)
	if err != nil {
		log.Println("error unmarshaling employee:", err)
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		c.JSON(http.StatusOK, &response)
		return
	}

	err = Validate.Struct(documentLetter)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		c.JSON(http.StatusOK, &response)
		return
	}

	rtn, err := db.Pool.Exec(
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

func GetLetterTypes(c *gin.Context) {
	var (
		documentTypes []models.DocumentType
		response      = models.Response{
			Code:    http.StatusOK,
			Message: http.StatusText(http.StatusOK),
			Time:    time.Now(),
		}
	)

	rows, err := db.Pool.Query(
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
		documentType := models.DocumentType{}
		err = rows.Scan(
			&documentType.Id,
			&documentType.Type,
		)

		documentTypes = append(documentTypes, documentType)
	}

	response.Payload = documentTypes

	c.JSON(http.StatusOK, &response)
}
