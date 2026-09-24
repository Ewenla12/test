package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/issachar/student-management-api/internal/models"
)

type SubjectHandler struct {
	DB *pgxpool.Pool
}

// GET /api/subjects — catalog, any authenticated party (admin or student)
func (h *SubjectHandler) Index(c *fiber.Ctx) error {
	rows, err := h.DB.Query(context.Background(), `SELECT id, name, code, created_at, updated_at FROM subjects ORDER BY name`)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	subjects := []models.Subject{}
	for rows.Next() {
		var s models.Subject
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		subjects = append(subjects, s)
	}

	return c.JSON(subjects)
}

// POST /api/admin/subjects
func (h *SubjectHandler) Store(c *fiber.Ctx) error {
	var req models.CreateSubjectRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" || req.Code == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "name and code are required"})
	}

	var s models.Subject
	err := h.DB.QueryRow(context.Background(),
		`INSERT INTO subjects (name, code) VALUES ($1, $2)
		 RETURNING id, name, code, created_at, updated_at`,
		req.Name, req.Code,
	).Scan(&s.ID, &s.Name, &s.Code, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "subject code already exists"})
	}

	return c.Status(fiber.StatusCreated).JSON(s)
}

// PUT/PATCH /api/admin/subjects/:id
func (h *SubjectHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req models.UpdateSubjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var s models.Subject
	err = h.DB.QueryRow(context.Background(), `
		UPDATE subjects SET
			name = COALESCE($1, name),
			code = COALESCE($2, code),
			updated_at = now()
		WHERE id = $3
		RETURNING id, name, code, created_at, updated_at`,
		req.Name, req.Code, id,
	).Scan(&s.ID, &s.Name, &s.Code, &s.CreatedAt, &s.UpdatedAt)

	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "subject not found"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(s)
}

// DELETE /api/admin/subjects/:id
func (h *SubjectHandler) Destroy(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tag, err := h.DB.Exec(context.Background(), `DELETE FROM subjects WHERE id = $1`, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "subject not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}
