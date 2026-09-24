package models

import "time"

type Student struct {
	ID                   int64     `json:"id"`
	IdentificationNumber string    `json:"identification_number"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	DateOfBirth          time.Time `json:"date_of_birth"`
	Class                *string   `json:"class"`
	Status               string    `json:"status"`
	Email                string    `json:"email"`
	PasswordHash         string    `json:"-"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Age is computed at read time, not stored — avoids the value going stale.
func (s Student) Age() int {
	now := time.Now()
	age := now.Year() - s.DateOfBirth.Year()
	if now.YearDay() < s.DateOfBirth.YearDay() {
		age--
	}
	return age
}

type StudentResponse struct {
	ID                   int64     `json:"id"`
	IdentificationNumber string    `json:"identification_number"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	DateOfBirth          string    `json:"date_of_birth"`
	Age                  int       `json:"age"`
	Class                *string   `json:"class"`
	Status               string    `json:"status"`
	Email                string    `json:"email"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (s Student) ToResponse() StudentResponse {
	return StudentResponse{
		ID:                   s.ID,
		IdentificationNumber: s.IdentificationNumber,
		FirstName:            s.FirstName,
		LastName:             s.LastName,
		DateOfBirth:          s.DateOfBirth.Format("2006-01-02"),
		Age:                  s.Age(),
		Class:                s.Class,
		Status:               s.Status,
		Email:                s.Email,
		CreatedAt:            s.CreatedAt,
		UpdatedAt:            s.UpdatedAt,
	}
}

// CreateStudentRequest is what the admin sends to create a student.
// identification_number is deliberately absent — it is server-generated.
type CreateStudentRequest struct {
	FirstName   string  `json:"first_name" validate:"required,max=100"`
	LastName    string  `json:"last_name" validate:"required,max=100"`
	DateOfBirth string  `json:"date_of_birth" validate:"required"` // "2006-01-02"
	Class       *string `json:"class"`
	Status      string  `json:"status"` // optional, defaults to "active"
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=8"`
}

type UpdateStudentRequest struct {
	FirstName   *string `json:"first_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth *string `json:"date_of_birth"`
	Class       *string `json:"class"`
	Status      *string `json:"status"`
	Email       *string `json:"email"`
	Password    *string `json:"password"`
}

type StudentFilter struct {
	Class     string
	Age       *int
	FirstName string
	LastName  string
	Status    string
	Page      int
	PerPage   int
}
