package app_errors

import (
	"time"
)

var (
	ErrFileAccess          = "an error occurred while accessing the file. Please try again later"
	ErrDataInsertion       = "an error occurred while inserting data into the database"
	ErrExcelOpen           = "an error occurred while opening the Excel file. Please try again later"
	ErrInvalidRequestBody  = "invalid request body"
	ErrInvalidCredentials  = "invalid credentials"
	ErrLogin               = "an error occurred while logging in. Please try again later"
	ErrInvalidtoken        = "invalid token"
	ErrInvalidRefreshtoken = "invalid refresh token"
	ErrTokenExpired        = "token expired"
	ErrRateLimitExceeded   = "rate limit exceeded"
	Errforbidden           = "you do not have permission to access this resource. Please contact the administrator for assistance"
	ErrPubsubMessage       = "error processing message from Pub/Sub subscription"
	ErrCacheClear          = "an error occurred while clearing the Redis cache. Please try again later"
	ErrDbRetrieve          = "error retrieving employee details from database"
	ErrDbInsert            = "an error occurred while performing the insert operation"
	ErrDataDeletion        = "an error occurred while deleting the record. Please try again later"
)

type ErrorTemplate struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Endpoint    string    `json:"endpoint"`
	Status_code int       `json:"status_code"`
	// Error_code  int       `json:"error_code"`
}

// logDetails := gin.H{
// 	"timestamp":    time.Now().UTC().Format(time.RFC3339),
// 	"level":        "INFO",
// 	"message":      "User successfully authenticated",
// 	"user_id":      userID,
// 	"endpoint":     c.Request.URL.Path,
// 	"status_code":  http.StatusOK,
// }
