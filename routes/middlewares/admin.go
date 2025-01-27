package middlewares

import (
	"github.com/gofiber/fiber/v2"
	"github.com/nusabangkit/finex/controllers/helpers"
	"github.com/nusabangkit/finex/models"
)

func AdminVaildator(c *fiber.Ctx) error {
	CurrentUser := c.Locals("CurrentUser").(*models.Member)

	if CurrentUser.Role != "admin" && CurrentUser.Role != "superadmin" {
		return c.Status(422).JSON(helpers.Errors{
			Errors: []string{"authz.invalid_permission"},
		})
	}

	return c.Next()
}
