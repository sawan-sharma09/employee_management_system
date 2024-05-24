package services

import (
	"database/sql"
	"fmt"
	"managedata/app_errors"
	grpcServices "managedata/grpc_services/grpc_client"
	initpack "managedata/init_pack"

	"net/http"
	"time"

	redisops "managedata/redisOps"
	"managedata/util"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func Appstart(c *gin.Context) {
	c.String(http.StatusOK, "Gin Application working")
}

func GetSingleEmployee(c *gin.Context) {
	id := c.Param("id")
	var emp util.Employee
	err := initpack.DbConn.QueryRow("SELECT * FROM employee_details WHERE id=?", id).Scan(&emp.ID, &emp.Name, &emp.Department, &emp.Salary)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Employee not found with id: ", id)
			logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: "Employee not found with id: " + id, Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
			c.JSON(http.StatusInternalServerError, logDetails)
			return
		} else {
			fmt.Println("Error in Mysql Select operation --> " + err.Error())
			logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrDbRetrieve, Endpoint: c.Request.URL.Path, Status_code: http.StatusNotFound}
			c.JSON(http.StatusNotFound, logDetails)
			return
		}
	}

	fmt.Printf("Fetched employee details: %+v\n", emp)

	c.JSON(http.StatusOK, gin.H{
		"Data": emp,
	})
}

func CreateEmployee(c *gin.Context) {

	var emp util.Employee

	err := c.ShouldBindJSON(&emp)
	if err != nil {
		fmt.Println("error in json bind of CreateEmployee route : ", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrInvalidRequestBody, Endpoint: c.Request.URL.Path, Status_code: http.StatusBadRequest}
		c.JSON(http.StatusBadRequest, logDetails)
		return
	}

	v := validator.New()

	if validationErr := v.Struct(emp); validationErr != nil {

		var validationErrors []string
		var failedLog string // this variable has been created to log all the validation error in server terminal one by one

		for _, e := range validationErr.(validator.ValidationErrors) {
			failedLog = fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag())
			fmt.Println(failedLog) //
			validationErrors = append(validationErrors, failedLog)
		}

		logDetails := map[string]interface{}{"Timestamp": time.Now(), "Level": "WARNING", "Message": "Invalid request body", "Endpoint": c.Request.URL.Path, "Errors": validationErrors, "Status_code": http.StatusBadRequest}
		c.JSON(http.StatusBadRequest, logDetails)
		return
	}

	// Insert the new employee into the Mysql database
	_, conn_err := initpack.DbConn.Exec("INSERT INTO employee_details (id,name, department, salary) VALUES (? ,?, ?, ?)", emp.ID, emp.Name, emp.Department, emp.Salary)
	if conn_err != nil {
		fmt.Println("error in Mysql insert operation", conn_err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrDataInsertion + conn_err.Error(), Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
		c.JSON(http.StatusInternalServerError, logDetails)
		return
	} else {
		fmt.Println("Employee created Successfully :", emp)
	}

	//Set data into redis
	redisops.RedisSet(emp)

	c.String(http.StatusCreated, "Employee created with ID: "+fmt.Sprint(emp.ID))
}

func DeleteEmployee(c *gin.Context) {

	id := c.Param("id")

	redisKey := fmt.Sprintf("employee:%s", id)
	fmt.Println("RedisKey: ", redisKey)

	//Check if the key for the Employee id exists or not
	if keyErr := redisops.RedisKeyExists(redisKey); keyErr != nil {
		fmt.Println(keyErr)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: keyErr.Error(), Endpoint: c.Request.URL.Path, Status_code: http.StatusNotFound}
		c.JSON(http.StatusNotFound, logDetails)
		return
	}

	// Delete the employee data from the  Mysql database
	_, deleteErr := initpack.DbConn.Exec("DELETE FROM employee_details WHERE id=?", id)
	if deleteErr != nil {
		fmt.Println("Error in Mysql delete operation : ", deleteErr)

		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrDataDeletion, Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
		c.JSON(http.StatusInternalServerError, logDetails)
		return
	}

	// Delete the employee data from redis
	redisDelErr := redisops.RedisDel(redisKey)
	if redisDelErr != nil {
		fmt.Println("err in Redis delete operation: ", redisDelErr)
		return
	}

	fmt.Printf("EmployeeId: %v deleted\n ", id)

	c.String(http.StatusOK, "Employee deleted successfully")
}

func UpdateEmployeeNew(c *gin.Context) {
	var emp util.Employee

	err := c.ShouldBindJSON(&emp)
	if err != nil {
		fmt.Println("Error in json bind of UpdateEmployee", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrInvalidRequestBody, Endpoint: c.Request.URL.Path, Status_code: http.StatusBadRequest}
		c.JSON(http.StatusBadRequest, logDetails)
		return
	}

	streamErr := grpcServices.Bidi_stream(emp)
	if streamErr != nil {
		fmt.Println("Error in Grpc Service: ", streamErr)
		return
	} else {
		c.String(http.StatusOK, "Employee updated successfully")
	}
}
