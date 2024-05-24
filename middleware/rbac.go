package middleware

import (
	"fmt"
	"managedata/app_errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	Role_Super_Admin = "super_admin"
	Role_Admin       = "admin"
	Role_User        = "user"
)

// Middleware to check if the user's role has the required permission
func RBACMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user's role has the required permission
		userRole := c.GetString("user_role")
		fmt.Println("User role in Rbac: ", userRole)
		if hasPermission(userRole, requiredRoles) {
			return
		}

		// If the user does not have permission, return a Forbidden response
		fmt.Println("User not authorized")
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "WARNING", Message: app_errors.Errforbidden, Endpoint: c.Request.URL.Path, Status_code: http.StatusForbidden}
		c.AbortWithStatusJSON(http.StatusForbidden, logDetails)
	}
}

// Check if the role has the required permission
func hasPermission(role string, requiredRoles []string) bool {
	// switch role {
	// case Role_Super_Admin, Role_Admin:
	// 	return true
	// default:
	// 	//return false for Role_User
	// 	return false
	// }

	for _, r := range requiredRoles {
		if role == r {
			return true
		}
	}
	return false
}
