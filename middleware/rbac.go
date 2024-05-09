package middleware

import (
	"fmt"
	"net/http"
)

const (
	Role_Super_Admin = "super_admin"
	Role_Admin       = "admin"
	Role_User        = "user"
)

// Middleware to check if the user's role has the required permission
func RBACMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check if the user's role has the required permission
		fmt.Println("User role -->", User_role)
		if hasPermission(User_role) {
			next.ServeHTTP(w, r)
			return
		}

		// If the user does not have permission, return a Forbidden response
		http.Error(w, "You do not have permission to access this resource. Please contact the administrator for assistance.", http.StatusForbidden)
	})
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
