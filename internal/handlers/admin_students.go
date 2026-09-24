package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/issachar/student-management-api/internal/models"
)

type StudentHandler struct {
	DB *pgxpool.Pool
}

// GET /api/admin/students?class=&age=&first_name=&last_name=&status=&page=&per_page=
func (h *StudentHandler) Index(c *fiber.Ctx) error {
	filter := models.StudentFilter{
		Class:     c.Query("class"),
		FirstName: c.Query("first_name"),
		LastName:  c.Query("last_name"),
		Status:    c.Query("status"),
		Page:      c.QueryInt("page", 1),
		PerPage:   c.QueryInt("per_page", 20),
	}
	if ageStr := c.Query("age"); ageStr != "" {
		if age, err := strconv.Atoi(ageStr); err == nil {
			filter.Age = &age
		}
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	where := []string{"1=1"}
	args := []interface{}{}
	argN := 1

	if filter.Class != "" {
		where = append(where, fmt.Sprintf("class = $%d", argN))
		args = append(args, filter.Class)
		argN++
	}
	if filter.FirstName != "" {
		where = append(where, fmt.Sprintf("first_name ILIKE $%d", argN))
		args = append(args, "%"+filter.FirstName+"%")
		argN++
	}
	if filter.LastName != "" {
		where = append(where, fmt.Sprintf("last_name ILIKE $%d", argN))
		args = append(args, "%"+filter.LastName+"%")
		argN++
	}
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argN))
		args = append(args, filter.Status)
		argN++
	}
	if filter.Age != nil {
		where = append(where, fmt.Sprintf("DATE_PART('year', AGE(date_of_birth)) = $%d", argN))
		args = append(args, *filter.Age)
		argN++
	}

	offset := (filter.Page - 1) * filter.PerPage
	query := fmt.Sprintf(`
		SELECT id, identification_number, first_name, last_name, date_of_birth,
		       class, status, email, created_at, updated_at
		FROM students
		WHERE %s
		ORDER BY id
		LIMIT $%d OFFSET $%d`,
		strings.Join(where, " AND "), argN, argN+1)
	args = append(args, filter.PerPage, offset)

	rows, err := h.DB.Query(context.Background(), query, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	students := []models.StudentResponse{}
	for rows.Next() {
		var s models.Student
		if err := rows.Scan(&s.ID, &s.IdentificationNumber, &s.FirstName, &s.LastName,
			&s.DateOfBirth, &s.Class, &s.Status, &s.Email, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		students = append(students, s.ToResponse())
	}

	return c.JSON(fiber.Map{
		"data": students,
		"page": filter.Page,
		"per_page": filter.PerPage,
	})
}

// POST /api/admin/students
// Request: CreateStudentRequest -> Response: StudentResponse (201)
func (h *StudentHandler) Store(c *fiber.Ctx) error {
	var req models.CreateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}
	if req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" || req.Email == "" || len(req.Password) < 8 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "missing or invalid required fields"})
	}

	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "date_of_birth must be YYYY-MM-DD"})
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
	}

	ctx := context.Background()
	tx, err := h.DB.Begin(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer tx.Rollback(ctx)

	idNumber, err := generateIdentificationNumber(ctx, tx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var s models.Student
	err = tx.QueryRow(ctx, `
		INSERT INTO students (identification_number, first_name, last_name, date_of_birth, class, status, email, password_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, identification_number, first_name, last_name, date_of_birth, class, status, email, created_at, updated_at`,
		idNumber, req.FirstName, req.LastName, dob, req.Class, status, req.Email, string(hash),
	).Scan(&s.ID, &s.IdentificationNumber, &s.FirstName, &s.LastName, &s.DateOfBirth,
		&s.Class, &s.Status, &s.Email, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "email already in use or invalid data: " + err.Error()})
	}

	if err := tx.Commit(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(s.ToResponse())
}

// GET /api/admin/students/:id
func (h *StudentHandler) Show(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var s models.Student
	err = h.DB.QueryRow(context.Background(), `
		SELECT id, identification_number, first_name, last_name, date_of_birth,
		       class, status, email, created_at, updated_at
		FROM students WHERE id = $1`, id,
	).Scan(&s.ID, &s.IdentificationNumber, &s.FirstName, &s.LastName, &s.DateOfBirth,
		&s.Class, &s.Status, &s.Email, &s.CreatedAt, &s.UpdatedAt)

	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student not found"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(s.ToResponse())
}

// PUT/PATCH /api/admin/students/:id
func (h *StudentHandler) Update(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	var req models.UpdateStudentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	sets := []string{}
	args := []interface{}{}
	argN := 1

	addSet := func(col string, val interface{}) {
		sets = append(sets, fmt.Sprintf("%s = $%d", col, argN))
		args = append(args, val)
		argN++
	}

	if req.FirstName != nil {
		addSet("first_name", *req.FirstName)
	}
	if req.LastName != nil {
		addSet("last_name", *req.LastName)
	}
	if req.DateOfBirth != nil {
		dob, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "date_of_birth must be YYYY-MM-DD"})
		}
		addSet("date_of_birth", dob)
	}
	if req.Class != nil {
		addSet("class", *req.Class)
	}
	if req.Status != nil {
		addSet("status", *req.Status)
	}
	if req.Email != nil {
		addSet("email", *req.Email)
	}
	if req.Password != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not hash password"})
		}
		addSet("password_hash", string(hash))
	}

	if len(sets) == 0 {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "no fields to update"})
	}

	sets = append(sets, fmt.Sprintf("updated_at = $%d", argN))
	args = append(args, time.Now())
	argN++

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE students SET %s WHERE id = $%d
		RETURNING id, identification_number, first_name, last_name, date_of_birth, class, status, email, created_at, updated_at`,
		strings.Join(sets, ", "), argN)

	var s models.Student
	err = h.DB.QueryRow(context.Background(), query, args...).Scan(
		&s.ID, &s.IdentificationNumber, &s.FirstName, &s.LastName, &s.DateOfBirth,
		&s.Class, &s.Status, &s.Email, &s.CreatedAt, &s.UpdatedAt)

	if err == pgx.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student not found"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(s.ToResponse())
}

// DELETE /api/admin/students/:id
func (h *StudentHandler) Destroy(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}

	tag, err := h.DB.Exec(context.Background(), `DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "student not found"})
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// generateIdentificationNumber atomically claims the next sequence number
// for the current year via id_sequences, formatted as STU-2026-00001.
func generateIdentificationNumber(ctx context.Context, tx pgx.Tx) (string, error) {
	year := time.Now().Year()

	var counter int
	err := tx.QueryRow(ctx, `
		INSERT INTO id_sequences (year, counter) VALUES ($1, 1)
		ON CONFLICT (year) DO UPDATE SET counter = id_sequences.counter + 1
		RETURNING counter`, year,
	).Scan(&counter)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("STU-%d-%05d", year, counter), nil
}
