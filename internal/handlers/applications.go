package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/issachar/student-management-api/internal/models"
)

type ApplicationHandler struct {
	DB *pgxpool.Pool
}

// GET /api/admin/subject-applications?status=pending
func (h *ApplicationHandler) Index(c *fiber.Ctx) error {
	status := c.Query("status")

	query := `
		SELECT ss.id, s.id, s.first_name || ' ' || s.last_name, s.identification_number,
		       sub.id, sub.name, ss.status, ss.applied_at, ss.reviewed_at, ss.reviewed_by
		FROM student_subject ss
		JOIN students s ON s.id = ss.student_id
		JOIN subjects sub ON sub.id = ss.subject_id`
	args := []interface{}{}

	if status != "" {
		query += ` WHERE ss.status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY ss.applied_at`

	rows, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	applications := []models.Application{}
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(&a.ID, &a.StudentID, &a.StudentName, &a.IdentificationNumber,
			&a.SubjectID, &a.SubjectName, &a.Status, &a.AppliedAt, &a.ReviewedAt, &a.ReviewedBy); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		applications = append(applications, a)
	}

	return c.JSON(applications)
}

// PATCH /api/admin/subject-applications/:id/approve
func (h *ApplicationHandler) Approve(c *fiber.Ctx) error {
	return h.review(c, "approved")
}

// PATCH /api/admin/subject-applications/:id/reject
func (h *ApplicationHandler) Reject(c *fiber.Ctx) error {
	return h.review(c, "rejected")
}

func (h *ApplicationHandler) review(c *fiber.Ctx, newStatus string) error {
	appID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid application id"})
	}
	adminID := c.Locals("user_id").(int64)

	tag, err := h.DB.Exec(context.Background(), `
		UPDATE student_subject
		SET status = $1, reviewed_at = $2, reviewed_by = $3, updated_at = $2
		WHERE id = $4 AND status = 'pending'`,
		newStatus, time.Now(), adminID, appID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no pending application with that id"})
	}

	return c.JSON(fiber.Map{"message": "application " + newStatus})
}
