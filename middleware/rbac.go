package middleware

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

const (
	Role_Super_Admin = "super_admin"
	Role_Admin       = "admin"
	Role_User        = "user"
)

// Middleware to check if the user's role has the required permission
func RBACMiddleware(c *fiber.Ctx) error {

	// Check if the user's role has the required permission
	fmt.Println("User role -->", User_role)
	if hasPermission(User_role) {
		return c.Next()
	}

	// If the user does not have permission, return a Forbidden response
	return c.Status(fiber.StatusForbidden).SendString("You do not have permission to access this resource. Please contact the administrator for assistance.")
}

// Check if the role has the required permission
func hasPermission(role string) bool {
	switch role {
	case Role_Super_Admin, Role_Admin:
		return true
	default:
		//return false for Role_User
		return false
	}
}
