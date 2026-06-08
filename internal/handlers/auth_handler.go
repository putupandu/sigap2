package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/models"
	"github.com/sigap2/sigap2/internal/services"
)

func ShowLogin(c *fiber.Ctx) error {
	return c.Render("auth/login", fiber.Map{}, "layouts/base")
}

func ProcessLogin(c *fiber.Ctx) error {
	email := c.FormValue("email")
	password := c.FormValue("password")

	token, err := services.LoginUser(email, password)
	if err != nil {
		return c.Render("auth/login", fiber.Map{
			"Error": "Invalid credentials",
		}, "layouts/base")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    token,
		HTTPOnly: true,
	})

	return c.Redirect("/dashboard")
}

func ShowRegister(c *fiber.Ctx) error {
	return c.Render("auth/register", fiber.Map{}, "layouts/base")
}

func ProcessRegister(c *fiber.Ctx) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	password := c.FormValue("password")
	role := models.RoleKorban // Force role to korban for public registration
	phone := c.FormValue("phone")

	err := services.RegisterUser(name, email, password, role, phone)
	if err != nil {
		return c.Render("auth/register", fiber.Map{
			"Error": err.Error(),
		}, "layouts/base")
	}

	return c.Redirect("/login?registered=1")
}

func Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "jwt",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})
	return c.Redirect("/")
}
