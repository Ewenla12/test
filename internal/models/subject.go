package models

import "time"

type Subject struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSubjectRequest struct {
	Name string `json:"name" validate:"required,max=150"`
	Code string `json:"code" validate:"required,max=20"`
}

type UpdateSubjectRequest struct {
	Name *string `json:"name"`
	Code *string `json:"code"`
}

// SubjectWithApplication is what a student sees for "my subjects" —
// the subject plus their own application status against it.
type SubjectWithApplication struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	Code              string     `json:"code"`
	ApplicationStatus string     `json:"application_status"`
	AppliedAt         time.Time  `json:"applied_at"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
}

// Application is a row in student_subject, joined with student + subject
// summaries for the admin review queue.
type Application struct {
	ID                   int64      `json:"id"`
	StudentID            int64      `json:"student_id"`
	StudentName          string     `json:"student_name"`
	IdentificationNumber string     `json:"identification_number"`
	SubjectID            int64      `json:"subject_id"`
	SubjectName          string     `json:"subject_name"`
	Status               string     `json:"status"`
	AppliedAt            time.Time  `json:"applied_at"`
	ReviewedAt           *time.Time `json:"reviewed_at"`
	ReviewedBy           *int64     `json:"reviewed_by"`
}
