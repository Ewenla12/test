package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/issachar/student-management-api/internal/handlers"
	"github.com/issachar/student-management-api/internal/middleware"
)

func Register(app *fiber.App, db *pgxpool.Pool, jwtSecret string) {
	adminAuth := &handlers.AdminAuthHandler{DB: db, JWTSecret: jwtSecret}
	studentAuth := &handlers.StudentAuthHandler{DB: db, JWTSecret: jwtSecret}
	students := &handlers.StudentHandler{DB: db}
	subjects := &handlers.SubjectHandler{DB: db}
	studentSubjects := &handlers.StudentSubjectHandler{DB: db}
	applications := &handlers.ApplicationHandler{DB: db}

	api := app.Group("/api")

	// Public
	api.Post("/admin/login", adminAuth.Login)
	api.Post("/student/login", studentAuth.Login)

	// Shared: subject catalog, any authenticated party
	api.Get("/subjects", middleware.RequireAuth(jwtSecret), subjects.Index)

	// Admin-only
	admin := api.Group("/admin", middleware.RequireAuth(jwtSecret), middleware.RequireRole("admin"))
	admin.Get("/students", students.Index)
	admin.Post("/students", students.Store)
	admin.Get("/students/:id", students.Show)
	admin.Put("/students/:id", students.Update)
	admin.Patch("/students/:id", students.Update)
	admin.Delete("/students/:id", students.Destroy)

	admin.Post("/subjects", subjects.Store)
	admin.Put("/subjects/:id", subjects.Update)
	admin.Patch("/subjects/:id", subjects.Update)
	admin.Delete("/subjects/:id", subjects.Destroy)

	admin.Get("/subject-applications", applications.Index)
	admin.Patch("/subject-applications/:id/approve", applications.Approve)
	admin.Patch("/subject-applications/:id/reject", applications.Reject)

	// Student-only
	student := api.Group("/student", middleware.RequireAuth(jwtSecret), middleware.RequireRole("student"))
	student.Get("/subjects", studentSubjects.Mine)
	student.Post("/subjects/:subjectId/apply", studentSubjects.Apply)
}
