package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
	"github.com/sigap2/sigap2/internal/services"
)

func ListRelawan(c *fiber.Ctx) error {
	var volunteers []models.User
	database.DB.Where("role = ?", models.RoleRelawan).Order("created_at desc").Find(&volunteers)

	return c.Render("volunteers/index", fiber.Map{
		"Volunteers": volunteers,
		"Success":    c.Query("success") == "1",
		"Deleted":    c.Query("deleted") == "1",
	}, "layouts/base")
}

func ShowCreateRelawan(c *fiber.Ctx) error {
	return c.Render("volunteers/create", fiber.Map{}, "layouts/base")
}

func CreateRelawan(c *fiber.Ctx) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	role := models.RoleRelawan // Hardcode to relawan
	phone := c.FormValue("phone")

	err := services.RegisterUser(name, email, password, role, phone)
	if err != nil {
		return c.Render("volunteers/create", fiber.Map{
			"Error": err.Error(),
			"Name": name,
			"Email": email,
			"Phone": phone,
		}, "layouts/base")
	}

	return c.Redirect("/volunteers?success=1")
}

func ProcessResetPassword(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	newPassword := c.FormValue("new_password")
	if newPassword == "" {
		return c.Status(400).SendString("Password cannot be empty")
	}

	err = services.UpdateUserPassword(uint(id), newPassword)
	if err != nil {
		return c.Status(500).SendString("Error updating password: " + err.Error())
	}

	return c.Redirect("/volunteers?success=1")
}

func ProcessChangeOwnPassword(c *fiber.Ctx) error {
	user := c.Locals("user").(models.User)
	
	oldPassword := c.FormValue("old_password")
	newPassword := c.FormValue("new_password")

	if oldPassword == "" || newPassword == "" {
		return c.Status(400).SendString("Password tidak boleh kosong")
	}

	err := services.ChangeOwnPassword(user.ID, oldPassword, newPassword)
	if err != nil {
		return c.Status(400).SendString(err.Error())
	}

	return c.Redirect(c.Get("Referer", "/") + "?pwd_success=1")
}

// GetVolunteerPassword returns the plain password for admin viewing
func GetVolunteerPassword(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid ID"})
	}

	var user models.User
	if err := database.DB.First(&user, uint(id)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "User not found"})
	}

	if user.PlainPassword == "" {
		return c.JSON(fiber.Map{"password": "(tidak tersedia - dibuat sebelum fitur ini)"})
	}

	return c.JSON(fiber.Map{"password": user.PlainPassword})
}

// DeleteVolunteer deletes a volunteer after verifying the confirmation name
func DeleteVolunteer(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).SendString("Invalid ID")
	}

	confirmName := c.FormValue("confirm_name")

	var user models.User
	if err := database.DB.First(&user, uint(id)).Error; err != nil {
		return c.Status(404).SendString("Relawan tidak ditemukan")
	}

	// 2-step verification: name must match
	if confirmName != user.Name {
		return c.Status(400).SendString("Nama konfirmasi tidak cocok")
	}

	// Soft delete the volunteer
	if err := database.DB.Delete(&user).Error; err != nil {
		return c.Status(500).SendString("Gagal menghapus relawan: " + err.Error())
	}

	return c.Redirect("/volunteers?deleted=1")
}
