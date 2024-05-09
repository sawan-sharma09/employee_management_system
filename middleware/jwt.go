package middleware

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	initpack "managedata/init_pack"
	"managedata/util"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Define JWT secret key
var (
	jwtKey    = []byte("S3cureJWT$ecretK3y!F0rD3m0App")
	User_role string // for role based authentication
)

// Define Claims struct for JWT
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Login handler to authenticate users and issue JWT tokens
func Login(w http.ResponseWriter, r *http.Request) {
	var credentials util.Credentials
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Check if the credentials are valid
	valid := isValidUser(credentials)
	if !valid {
		fmt.Println("Invalid Credentials...!!!")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Create JWT token
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: credentials.Username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println("Error generating JWT token:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set JWT token as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Cookie generated...!!"))
}

// Middleware function to authenticate JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from request (e.g., from Authorization header)
		tokenString := extractToken(r)

		// Parse and validate the token
		claims, err := parseToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Check token expiration
		if claims.ExpiresAt < time.Now().Unix() {
			http.Error(w, "Token has expired", http.StatusUnauthorized)
			return
		}

		// Proceed to the next handler if token is valid
		next.ServeHTTP(w, r)
	})
}

// Function to extract token from request (e.g., from Authorization header)
func extractToken(r *http.Request) string {
	// Extract token from Authorization header or other sources
	// Example: Authorization: Bearer <token>
	token := r.Header.Get("Authorization")
	if token != "" {
		parts := strings.Split(token, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}
	return ""
}

// Function to parse and validate token
func parseToken(tokenString string) (*Claims, error) {
	// Parse and validate token using jwt.ParseWithClaims
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

// Helper function to validate user credentials (replace this with your own authentication logic)
func isValidUser(credentials util.Credentials) bool {

	var realPassword string

	// Fetch password for authentication and if User is authenticated, use the 'User_role' for authorization
	err := initpack.Conn.QueryRow("SELECT password,role FROM cred_manager WHERE username=?", credentials.Username).Scan(&realPassword, &User_role)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User %s not found", credentials.Username)
			return false
		} else {
			fmt.Println("error in Mysql Select operation", err)
			return false
		}
	}
	fmt.Println("From valid user-->", User_role)
	return credentials.Password == realPassword
}
