package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	Role_Super_Admin = "super_admin"
	Role_Admin       = "admin"
	Role_User        = "user"
)

// Middleware to check if the user's role has the required permission
func RBACMiddleware(c *gin.Context) {

	// Check if the user's role has the required permission
	fmt.Println("User role -->", User_role)
	if hasPermission(User_role) {
		return
	}

	// If the user does not have permission, return a Forbidden response
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "You do not have permission to access this resource. Please contact the administrator for assistance."})
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
