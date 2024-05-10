package services

import (
	"database/sql"
	"fmt"
	grpcServices "managedata/grpc_services/grpc_client"
	initpack "managedata/init_pack"
	"net/http"

	redisops "managedata/redisOps"
	"managedata/util"

	"github.com/gin-gonic/gin"
)

func Appstart(c *gin.Context) {
	c.String(http.StatusOK, "Gin Application working")
}

func GetSingleEmployee(c *gin.Context) {
	id := c.Param("id")
	var emp util.Employee
	err := initpack.Conn.QueryRow("SELECT * FROM employee_details WHERE id=?", id).Scan(&emp.ID, &emp.Name, &emp.Department, &emp.Salary)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Employee not found with id: ", id)
			c.String(http.StatusInternalServerError, "Employee not found with id: "+id)
			return
		} else {
			c.String(http.StatusNotFound, "Error in Mysql Select operation --> "+err.Error())
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
		fmt.Println("error in NewDecoder: ", err)
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Insert the new employee into the Mysql database
	_, conn_err := initpack.Conn.Exec("INSERT INTO employee_details (id,name, department, salary) VALUES (? ,?, ?, ?)", emp.ID, emp.Name, emp.Department, emp.Salary)
	if conn_err != nil {
		fmt.Println("error in Mysql insert operation", conn_err)
		c.String(http.StatusInternalServerError, "MySql Insert Error: "+conn_err.Error())
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
		c.String(http.StatusNotFound, keyErr.Error())
		return
	}

	// Delete the employee data from the  Mysql database
	_, deleteErr := initpack.Conn.Exec("DELETE FROM employee_details WHERE id=?", id)
	if deleteErr != nil {
		fmt.Println("Error in Mysql delete operation : ", deleteErr)
		c.String(http.StatusNotFound, "Mysql Delete Error: "+deleteErr.Error())
		return
	}

	// Delete the employee data from redis
	redisDelErr := redisops.RedisDel(redisKey)
	if redisDelErr != nil {
		fmt.Println("err in Redis delete operation: ", redisDelErr)
		c.String(http.StatusInternalServerError, "Redis Delete Error: "+redisDelErr.Error())
		return
	}

	fmt.Printf("EmployeeId: %v deleted\n ", id)

	c.String(http.StatusOK, "Employee deleted successfully")
}

func UpdateEmployeeNew(c *gin.Context) {
	var emp util.Employee

	err := c.ShouldBindJSON(&emp)
	if err != nil {
		fmt.Println("Error in NewDecoder of UpdateEmployee", err)
		c.String(http.StatusBadRequest, "Invalid request body: "+err.Error())
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
