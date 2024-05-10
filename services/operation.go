package services

import (
	"database/sql"
	"fmt"
	grpcServices "managedata/grpc_services/grpc_client"
	initpack "managedata/init_pack"

	redisops "managedata/redisOps"
	"managedata/util"

	"github.com/gofiber/fiber/v2"
)

func GetSingleEmployee(c *fiber.Ctx) error {
	id := c.Params("id")
	var emp util.Employee
	err := initpack.Conn.QueryRow("SELECT * FROM employee_details WHERE id=?", id).Scan(&emp.ID, &emp.Name, &emp.Department, &emp.Salary)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusInternalServerError).SendString("Employee not found")
		} else {
			return c.Status(fiber.StatusNotFound).SendString("Error in Mysql Select operation --> " + err.Error())
		}
	}

	fmt.Printf("Fetched employee details: %+v\n", emp)
	responseData := map[string]interface{}{
		"Data": emp,
	}
	return c.JSON(&responseData)
}

func Appstart(c *fiber.Ctx) error {
	return c.SendString("Fiber Golang Working")
}

func CreateEmployee(c *fiber.Ctx) error {

	var emp util.Employee

	err := c.BodyParser(&emp)
	if err != nil {
		fmt.Println("error in NewDecoder: ", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body: " + err.Error())
	}

	// Insert the new employee into the Mysql database
	_, conn_err := initpack.Conn.Exec("INSERT INTO employee_details (id,name, department, salary) VALUES (? ,?, ?, ?)", emp.ID, emp.Name, emp.Department, emp.Salary)
	if conn_err != nil {
		fmt.Println("error in Mysql insert operation", conn_err)
		return c.Status(fiber.StatusInternalServerError).SendString("MySql Insert Error: " + conn_err.Error())
	} else {
		fmt.Println("Employee created Successfully :", emp)
	}

	//Set data into redis
	redisops.RedisSet(emp)

	return c.Status(fiber.StatusCreated).SendString("Employee created with ID: " + fmt.Sprint(emp.ID))
}

func DeleteEmployee(c *fiber.Ctx) error {

	id := c.Params("id")

	redisKey := fmt.Sprintf("employee:%s", id)

	//Check if the key for the Employee id exists or not
	if keyErr := redisops.RedisKeyExists(redisKey); keyErr != nil {
		fmt.Println(keyErr)
		return keyErr
	}

	// Delete the employee data from the  Mysql database
	_, deleteErr := initpack.Conn.Exec("DELETE FROM employee_details WHERE id=?", id)
	if deleteErr != nil {
		fmt.Println("Error in Mysql delete operation : ", deleteErr)
		return c.Status(fiber.StatusNotFound).SendString("Mysql Delete Error: " + deleteErr.Error())
	}

	// Delete the employee data from redis
	redisDelErr := redisops.RedisDel(redisKey)
	if redisDelErr != nil {
		fmt.Println("err in Redis delete operation: ", redisDelErr)
		return c.Status(fiber.StatusInternalServerError).SendString("Redis Delete Error: " + redisDelErr.Error())
	}

	fmt.Printf("EmployeeId: %v deleted\n ", id)

	return c.Status(fiber.StatusOK).SendString("Employee deleted successfully")
}

func UpdateEmployeeNew(c *fiber.Ctx) error {

	var emp util.Employee

	err := c.BodyParser(&emp)
	if err != nil {
		fmt.Println("Error in NewDecoder of UpdateEmployee", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body: " + err.Error())
	}

	streamErr := grpcServices.Bidi_stream(emp)
	if streamErr != nil {
		fmt.Println("Error in Grpc Service: ", streamErr)
		return streamErr
	} else {
		c.Status(fiber.StatusOK).SendString("Employee updated successfully")
	}

	return nil
}
