package models

import "time"

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

type LetterFilter struct {
	Name               string `json:"name" validate:"min=1"`
	Sender             string `json:"sender" validate:"min=1"`
	DepartmentTypeId   int    `json:"department_type_id" validate:"number,min=1"`
	RegistrationNumber string `json:"registration_number" validate:"min=1"`
	OutgoingNumber     string `json:"outgoing_number" validate:"min=1"`
	RowsLimit          uint   `json:"rows_limit" validate:"required,number,min=1"`
	RowsOffset         uint   `json:"rows_offset" validate:"number,min=0"`
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
