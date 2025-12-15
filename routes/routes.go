package routes

import (
	"github.com/Joko206/UAS_PWEB1/controllers"
	"github.com/gofiber/fiber/v2"
)

// Middleware untuk autentikasi
func AuthMiddleware(c *fiber.Ctx) error {
	_, err := controllers.Authenticate(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"data":    nil,
			"success": false,
			"message": "Unauthorized",
		})
	}
	return c.Next()
}

func Setup(app *fiber.App) {
	// Root Route
	start := app.Group("/")
	start.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello World")
	})

	// User Routes
	api := app.Group("/user")
	api.Get("/get-user", controllers.User)
	api.Post("/register", controllers.Register)
	api.Post("/login", controllers.Login)
	api.Get("/logout", controllers.Logout)

	// Kategori Routes (Only Admin)
	kategori := app.Group("/kategori", AuthMiddleware)
	kategori.Get("/get-kategori", controllers.GetKategori)
	kategori.Post("/add-kategori", controllers.RoleMiddleware([]string{"admin"}), controllers.AddKategori)
	kategori.Patch("/update-kategori/:id", controllers.RoleMiddleware([]string{"admin"}), controllers.UpdateKategori)
	kategori.Delete("/delete-kategori/:id", controllers.RoleMiddleware([]string{"admin"}), controllers.DeleteKategori)

	// Tingkatan Routes (Only Admin and Teacher)
	tingkatan := app.Group("/tingkatan", AuthMiddleware)
	tingkatan.Get("/get-tingkatan", controllers.GetTingkatan)
	tingkatan.Post("/add-tingkatan", controllers.RoleMiddleware([]string{"admin"}), controllers.AddTingkatan)
	tingkatan.Patch("/update-tingkatan/:id", controllers.RoleMiddleware([]string{"admin"}), controllers.UpdateTingkatan)
	tingkatan.Delete("/delete-tingkatan/:id", controllers.RoleMiddleware([]string{"admin"}), controllers.DeleteTingkatan)

	// Kelas Routes (Admin, Teacher, Student)
	kelas := app.Group("/kelas", AuthMiddleware)
	kelas.Get("/get-kelas", controllers.GetKelas)
	kelas.Post("/add-kelas", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.AddKelas)
	kelas.Patch("/update-kelas/:id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.UpdateKelas)
	kelas.Delete("/delete-kelas/:id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.DeleteKelas)
	kelas.Post("/join-kelas", controllers.JoinKelas)
	kelas.Post("/join-by-code", controllers.JoinKelasByCode)
	kelas.Get("/get-kelas-by-user", controllers.GetKelasByUserID)

	// Kuis Routes (Admin, Teacher)
	kuis := app.Group("/kuis", AuthMiddleware)
	kuis.Get("/get-kuis", controllers.GetKuis)
	kuis.Get("/get-all-kuis", controllers.RoleMiddleware([]string{"admin"}), controllers.GetAllKuis)
	kuis.Post("/add-kuis", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.AddKuis)
	kuis.Patch("/update-kuis/:id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.UpdateKuis)
	kuis.Delete("/delete-kuis/:id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.DeleteKuis)
	kuis.Get("/filter-kuis", controllers.FilterKuis)

	// Soal Routes (Admin, Teacher)
	soal := app.Group("/soal", AuthMiddleware)
	soal.Get("/get-soal", controllers.GetSoal)
	soal.Get("/get-soal/:kuis_id", controllers.GetSoalByKuisID)
	soal.Post("/add-soal", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.AddSoal)
	soal.Patch("/update-soal/:id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.UpdateSoal)
	soal.Delete("/delete-soal/:id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.DeleteSoal)

	// Pendidikan Routes (Only Admin)
	pendidikan := app.Group("/pendidikan", AuthMiddleware)
	pendidikan.Get("/get-pendidikan", controllers.GetPendidikan)
	pendidikan.Post("/add-pendidikan", controllers.RoleMiddleware([]string{"admin"}), controllers.AddPendidikan)
	pendidikan.Patch("/update-pendidikan/:id", controllers.RoleMiddleware([]string{"admin"}), controllers.UpdatePendidikan)
	pendidikan.Delete("/delete-pendidikan/:id", controllers.RoleMiddleware([]string{"admin"}), controllers.DeletePendidikan)

	// Hasil Kuis Routes (Admin, Teacher, Student)
	result := app.Group("/hasil-kuis", AuthMiddleware)
	result.Get("/my-results", controllers.GetAllHasilKuisByUser)
	result.Get("/user/:user_id", controllers.RoleMiddleware([]string{"admin", "teacher"}), controllers.GetHasilKuisByUserID)
	result.Post("/submit-jawaban", controllers.SubmitJawaban)
	result.Get("/:user_id/:kuis_id", controllers.GetHasilKuis)

	// Audit Routes (Admin only)
	audit := app.Group("/audit", AuthMiddleware)
	audit.Get("/logs", controllers.RoleMiddleware([]string{"admin"}), controllers.GetAuditLogs)

	// 404 Handler - must be last
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"data":    nil,
			"success": false,
			"message": "Endpoint not found",
			"path":    c.Path(),
			"method":  c.Method(),
		})
	})
}
