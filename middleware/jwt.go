package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"managedata/app_errors"
	initpack "managedata/init_pack"
	"managedata/util"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Define JWT secret key
var (
	jwtKey        = []byte("S3cureJWT$ecretK3y!F0rD3m0App")
	refreshJwtKey = []byte("R3freshJWT$ecretK3y!F0rD3m0App")
)

// Define Claims struct for JWT
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

type RefreshClaims struct {
	Username string
	Role     string `json:"role"`
	jwt.StandardClaims
}

// Login handler to authenticate users and issue JWT tokens
func Login(c *gin.Context) {
	var credentials util.Credentials

	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		fmt.Println("Invalid request body in Login: ", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrInvalidRequestBody, Endpoint: c.Request.URL.Path, Status_code: http.StatusBadRequest}
		c.JSON(http.StatusBadRequest, logDetails)
		return
	}

	// Check if the credentials are valid
	userRole, valid := isValidUser(credentials)
	if !valid {
		fmt.Println("Invalid Credentials...!!!")
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "WARNING", Message: app_errors.ErrInvalidCredentials, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.JSON(http.StatusUnauthorized, logDetails)
		return
	}

	// Create JWT access token
	expirationTime := time.Now().Add(10 * time.Minute)
	claims := &Claims{
		Username: credentials.Username,
		Role:     userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println("Error generating JWT token:", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrLogin, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.JSON(http.StatusUnauthorized, logDetails)
		return
	}

	//Create Refresh Token
	refreshExpirationTime := time.Now().Add(24 * time.Hour)
	refreshClaims := &RefreshClaims{
		Username: credentials.Username,
		Role:     userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: refreshExpirationTime.Unix(),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, refreshErr := refreshToken.SignedString(refreshJwtKey)
	if refreshErr != nil {
		fmt.Println("Error generating Refresh token: ", refreshErr)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrLogin, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.JSON(http.StatusUnauthorized, logDetails)
		return
	}

	//Set Cookies for access token and refresh token
	c.SetCookie("token", tokenString, int(time.Until(expirationTime).Seconds()), "/", "", false, true)
	c.SetCookie("refresh_token", refreshTokenString, int(time.Until(refreshExpirationTime).Seconds()), "/", "", false, true)

	c.String(http.StatusOK, "User Authenticated Successfully")
}

func Refresh(c *gin.Context) {
	refreshTokenString, err := c.Cookie("refresh_token")
	if err != nil {
		fmt.Println("Invalid refresh token :", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrInvalidRefreshtoken, Endpoint: c.Request.URL.Path, Status_code: http.StatusBadRequest}
		c.JSON(http.StatusBadRequest, logDetails)
		return
	}

	refreshClaims := &RefreshClaims{}
	refreshToken, claimsErr := jwt.ParseWithClaims(refreshTokenString, refreshClaims, func(token *jwt.Token) (interface{}, error) {
		return refreshJwtKey, nil
	})

	if claimsErr != nil || !refreshToken.Valid {
		fmt.Println("Invalid refresh token: ", claimsErr)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "WARNING", Message: app_errors.ErrInvalidtoken, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.JSON(http.StatusUnauthorized, logDetails)
		return
	}

	if refreshClaims.ExpiresAt < time.Now().Unix() {
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "INFO", Message: app_errors.ErrTokenExpired, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.JSON(http.StatusUnauthorized, logDetails)
		return
	}

	// Generate new access token
	expirationTime := time.Now().Add(10 * time.Minute)
	claims := &Claims{
		Username: refreshClaims.Username,
		Role:     refreshClaims.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		fmt.Println("Error in signing token string: ", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrLogin, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.JSON(http.StatusUnauthorized, logDetails)
		return
	}

	// Set new access token as cookie
	c.SetCookie("token", tokenString, int(time.Until(expirationTime).Seconds()), "/", "", false, true)

	fmt.Println("Token has been refreshed..")
	c.String(http.StatusOK, "Token Refreshed Successfully")

}

func AuthMiddleware(c *gin.Context) {
	tokenString := extractToken(c)

	// Parse and validate the token
	claims, err := parseToken(tokenString)
	if err != nil {
		fmt.Println("Token parse error in Authmiddleware: ", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "WARNING", Message: app_errors.ErrInvalidtoken, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.AbortWithStatusJSON(http.StatusUnauthorized, logDetails)
		return
	}

	// Check token expiration
	if claims.ExpiresAt < time.Now().Unix() {
		fmt.Println("Token expired")
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "INFO", Message: app_errors.ErrTokenExpired, Endpoint: c.Request.URL.Path, Status_code: http.StatusUnauthorized}
		c.AbortWithStatusJSON(http.StatusUnauthorized, logDetails)
		return
	}
	c.Set("user_role", claims.Role)
}

// Function to extract token from request (e.g., from Authorization header)
func extractToken(c *gin.Context) string {
	// Extract token from Authorization header or other sources
	// Example: Authorization: Bearer <token>
	token := c.GetHeader("Authorization")
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
func isValidUser(credentials util.Credentials) (string, bool) {

	var realPassword, userRole string

	// Prepare SQL statement to query the user credentials
	err := initpack.DbConn.QueryRow("SELECT password,role FROM cred_manager WHERE username=?", credentials.Username).Scan(&realPassword, &userRole)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("User %s not found", credentials.Username)
			return "", false
		} else {
			fmt.Println("error in Mysql Select operation", err)
			return "", false
		}
	}
	return userRole, credentials.Password == realPassword
}
