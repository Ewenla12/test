package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/issachar/student-management-api/internal/middleware"
	"github.com/issachar/student-management-api/internal/models"
)

type StudentAuthHandler struct {
	DB        *pgxpool.Pool
	JWTSecret string
}

// POST /api/student/login
// Request:  { "email": string, "password": string }
// Response: { "token": string, "student": StudentResponse }
func (h *StudentAuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var s models.Student
	err := h.DB.QueryRow(context.Background(), `
		SELECT id, identification_number, first_name, last_name, date_of_birth,
		       class, status, email, password_hash, created_at, updated_at
		FROM students WHERE email = $1`,
		req.Email,
	).Scan(&s.ID, &s.IdentificationNumber, &s.FirstName, &s.LastName, &s.DateOfBirth,
		&s.Class, &s.Status, &s.Email, &s.PasswordHash, &s.CreatedAt, &s.UpdatedAt)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(s.PasswordHash), []byte(req.Password)) != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := middleware.GenerateToken(h.JWTSecret, s.ID, "student")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token":   token,
		"student": s.ToResponse(),
	})
}
