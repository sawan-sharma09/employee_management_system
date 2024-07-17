package services

import (
	"context"
	"database/sql"
	"errors"
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

// @Tags			Health
// @Summary		Check if the application is running
// @Description	Check the aliveness of the application
// @Success		200	{string}	string	"Gin Application working"
// @Router			/ [get]
func Appstart(c *gin.Context) {
	c.String(http.StatusOK, "Gin Application working")
}

// @Tags			Employee
// @Summary		Get a single employee
// @Description	Fetch a record of a specific employee by ID
// @Param			id	path		int	true	"Employee ID"
// @Success		200	{object}	util.Employee
// @Failure		404	{object}	app_errors.ErrorTemplate "Employee not found"
// @Failure		500	{object}	app_errors.ErrorTemplate
// @Security		BearerAuth
// @Router			/v1/get_employee/{id} [get]
func GetSingleEmployee(c *gin.Context) {
	id := c.Param("id")
	var emp util.Employee
	err := initpack.PostgresPool.QueryRow(context.Background(), "SELECT * FROM employees WHERE id=$1", id).Scan(&emp.ID, &emp.Name, &emp.Department, &emp.Salary)
	fmt.Println("err---->", err)
	fmt.Println("Errnorows---->", sql.ErrNoRows)
	fmt.Println(errors.Is(err, sql.ErrNoRows))

	if err != nil {
		if err.Error() == app_errors.ErrEmployeeNotFound {
			fmt.Println("Employee not found with id: ", id)
			logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: "Employee not found with id: " + id, Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
			c.JSON(http.StatusNotFound, logDetails)
			return
		} else {
			fmt.Println("Error in Mysql Select operation --> " + err.Error())
			logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrDbRetrieve, Endpoint: c.Request.URL.Path, Status_code: http.StatusNotFound}
			c.JSON(http.StatusInternalServerError, logDetails)
			return
		}
	}

	fmt.Printf("Fetched employee details: %+v\n", emp)

	c.JSON(http.StatusOK, gin.H{
		"Data": emp,
	})
}

// @Tags			Employee
// @Summary		Create a new employee
// @Description	Create a new employee record
// @Accept			json
// @Produce		json
// @Param			employee	body		util.Employee	true	"Employee Data"
// @Success		201			{string}	string			"Employee created with ID: X"
// @Failure		400			{object}	app_errors.ErrorTemplate
// @Failure		500			{object}	app_errors.ErrorTemplate
// @Security		BearerAuth
// @Router			/v1/create_employee [post]
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
	_, conn_err := initpack.PostgresPool.Exec(context.Background(), "INSERT INTO employees (id,name, department, salary) VALUES ($1 ,$2, $3, $4)", emp.ID, emp.Name, emp.Department, emp.Salary)
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

//	@Tags			Employee
//	@Summary		Delete an employee
//	@Description	Delete an employee by ID
//	@Param			id	path		string	true	"Employee ID"
//	@Success		200	{string}	string	"Employee deleted successfully"
//	@Failure		404	{object}	app_errors.ErrorTemplate
//	@Failure		500	{object}	app_errors.ErrorTemplate
//
// @Security		BearerAuth
// @Router			/v1/delete_employee/{id} [delete]
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
	_, deleteErr := initpack.PostgresPool.Exec(context.Background(), "DELETE FROM employees WHERE id=$1", id)
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

//	@Tags			Employee
//	@Summary		Update an employee
//	@Description	Update an employee record
//	@Accept			json
//	@Produce		json
//	@Param			employee	body		util.Employee	true	"Employee Data"
//	@Success		200			{string}	string			"Employee updated successfully"
//	@Failure		400			{object}	app_errors.ErrorTemplate
//
// @Security		BearerAuth
// @Router			/v1/update_employee [put]
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
