package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sigap2/sigap2/internal/config"
	"github.com/sigap2/sigap2/internal/database"
	"github.com/sigap2/sigap2/internal/models"
)

func RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from cookie
		tokenString := c.Cookies("jwt")
		if tokenString == "" {
			return c.Redirect("/login")
		}

		// Parse token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			// Clear invalid cookie
			c.Cookie(&fiber.Cookie{
				Name:     "jwt",
				Value:    "",
				Expires:  time.Now().Add(-time.Hour),
				HTTPOnly: true,
			})
			return c.Redirect("/login")
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Redirect("/login")
		}

		// Check expiry
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				return c.Redirect("/login")
			}
		}

		// Fetch user from DB
		var user models.User
		userID := uint(claims["sub"].(float64))
		database.DB.First(&user, userID)

		if user.ID == 0 {
			return c.Redirect("/login")
		}

		// Set user in locals
		c.Locals("user", user)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(models.User)
		if !ok {
			return c.Redirect("/login")
		}

		hasRole := false
		for _, role := range roles {
			if user.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).SendString("Forbidden: You don't have access to this resource. Your role is: " + user.Role)
		}

		return c.Next()
	}
}

func InjectUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("jwt")
		if tokenString != "" {
			token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(config.AppConfig.JWTSecret), nil
			})

			if token != nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					var user models.User
					userID := uint(claims["sub"].(float64))
					database.DB.First(&user, userID)
					if user.ID != 0 {
						c.Locals("user", user)
						// also pass to template
						c.Bind(fiber.Map{
							"User": user,
						})
					}
				}
			}
		}
		return c.Next()
	}
}
