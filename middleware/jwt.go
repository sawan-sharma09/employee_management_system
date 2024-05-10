package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	initpack "managedata/init_pack"
	"managedata/util"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber/v2"
)

// Define JWT secret key
var (
	jwtKey    = []byte("S3cureJWT$ecretK3y!F0rD3m0App")
	User_role string
)

// Define Claims struct for JWT
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// Login handler to authenticate users and issue JWT tokens
func Login(c *fiber.Ctx) error {
	var credentials util.Credentials

	err := c.BodyParser(&credentials)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Check if the credentials are valid
	valid := isValidUser(credentials)
	if !valid {
		fmt.Println("Invalid Credentials...!!!")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials !!",
		})
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
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	// Set JWT token as a cookie
	c.Cookie(&fiber.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})
	return c.SendStatus(fiber.StatusOK)
}

func AuthMiddleware(c *fiber.Ctx) error {
	tokenString := extractToken(c)

	// Parse and validate the token
	claims, err := parseToken(tokenString)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).SendString("Invalid token")
	}

	// Check token expiration
	if claims.ExpiresAt < time.Now().Unix() {
		return c.Status(fiber.StatusUnauthorized).SendString("Token has expired")
	}

	// Proceed to the next handler if token is valid
	return c.Next()
}

// Function to extract token from request (e.g., from Authorization header)
func extractToken(c *fiber.Ctx) string {
	// Extract token from Authorization header or other sources
	// Example: Authorization: Bearer <token>
	token := c.Get("Authorization")
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

	// Prepare SQL statement to query the user credentials
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
	return credentials.Password == realPassword
}
