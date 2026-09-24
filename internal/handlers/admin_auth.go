package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/issachar/student-management-api/internal/middleware"
	"github.com/issachar/student-management-api/internal/models"
)

type AdminAuthHandler struct {
	DB        *pgxpool.Pool
	JWTSecret string
}

// POST /api/admin/login
// Request:  { "email": string, "password": string }
// Response: { "token": string, "admin": { id, name, email } }
func (h *AdminAuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	var admin models.Admin
	err := h.DB.QueryRow(context.Background(),
		`SELECT id, name, email, password_hash FROM admins WHERE email = $1`,
		req.Email,
	).Scan(&admin.ID, &admin.Name, &admin.Email, &admin.PasswordHash)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)) != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
	}

	token, err := middleware.GenerateToken(h.JWTSecret, admin.ID, "admin")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not generate token"})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"admin": fiber.Map{"id": admin.ID, "name": admin.Name, "email": admin.Email},
	})
}
