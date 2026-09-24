package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/issachar/student-management-api/internal/models"
)

type StudentSubjectHandler struct {
	DB *pgxpool.Pool
}

// GET /api/student/subjects?status=pending|approved|rejected
// Response: the authenticated student's subjects + their application status.
// With no ?status filter, returns every subject they've ever applied to.
func (h *StudentSubjectHandler) Mine(c *fiber.Ctx) error {
	studentID := c.Locals("user_id").(int64)
	status := c.Query("status")

	query := `
		SELECT sub.id, sub.name, sub.code, ss.status, ss.applied_at, ss.reviewed_at
		FROM student_subject ss
		JOIN subjects sub ON sub.id = ss.subject_id
		WHERE ss.student_id = $1`
	args := []interface{}{studentID}

	if status != "" {
		query += ` AND ss.status = $2`
		args = append(args, status)
	}
	query += ` ORDER BY ss.applied_at DESC`

	rows, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	results := []models.SubjectWithApplication{}
	for rows.Next() {
		var s models.SubjectWithApplication
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.ApplicationStatus, &s.AppliedAt, &s.ReviewedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		results = append(results, s)
	}

	return c.JSON(results)
}

// POST /api/student/subjects/:subjectId/apply
// Creates a pending application. Fails if one already exists for this pair,
// regardless of its current status (see note below on reapplying).
func (h *StudentSubjectHandler) Apply(c *fiber.Ctx) error {
	studentID := c.Locals("user_id").(int64)

	subjectID, err := c.ParamsInt("subjectId")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid subject id"})
	}

	var existingStatus string
	err = h.DB.QueryRow(context.Background(),
		`SELECT status FROM student_subject WHERE student_id = $1 AND subject_id = $2`,
		studentID, subjectID,
	).Scan(&existingStatus)

	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": "you have already applied to this subject, current status: " + existingStatus,
		})
	} else if err != pgx.ErrNoRows {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	_, err = h.DB.Exec(context.Background(),
		`INSERT INTO student_subject (student_id, subject_id, status, applied_at)
		 VALUES ($1, $2, 'pending', now())`,
		studentID, subjectID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "application submitted, pending admin approval"})
}
