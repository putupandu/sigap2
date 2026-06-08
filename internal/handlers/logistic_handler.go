package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sigap2/sigap2/internal/models"
	"github.com/sigap2/sigap2/internal/services"
)

func ListLogistics(c *fiber.Ctx) error {
	logistics, _ := services.GetLogistics()
	return c.Render("logistics/index", fiber.Map{
		"Logistics": logistics,
	}, "layouts/base")
}

func ShowLogisticForm(c *fiber.Ctx) error {
	return c.Render("logistics/form", fiber.Map{}, "layouts/base")
}

func CreateLogistic(c *fiber.Ctx) error {
	qty, _ := strconv.Atoi(c.FormValue("quantity"))
	item := models.Logistic{
		ItemName: c.FormValue("item_name"),
		Quantity: qty,
		Unit:     c.FormValue("unit"),
	}

	services.CreateLogistic(item)
	return c.Redirect("/logistics")
}

func DeleteLogistic(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err == nil && id > 0 {
		services.DeleteLogistic(uint(id))
	}
	return c.Redirect("/logistics")
}
