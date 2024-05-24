package db

import (
	"database/sql"
	"errors"
	"fmt"
	"grpc_new_server/conn"
	"grpc_new_server/grpc_services/utils"
)

func DbUpdateEmp(emp *utils.Employee) (utils.Employee, error) {

	var fetchedData utils.Employee

	//first check the employee exists or not
	selectErr := conn.DbConnect.QueryRow("SELECT id,name,department,salary FROM employee_details WHERE id=?", emp.ID).Scan(&fetchedData.ID, &fetchedData.Name, &fetchedData.Department, &fetchedData.Salary)
	if selectErr != nil {
		if selectErr == sql.ErrNoRows {
			return utils.Employee{}, errors.New("Employee not found with id: " + fmt.Sprint(emp.ID))
		} else {
			return utils.Employee{}, errors.New("Error in Mysql Select operation" + selectErr.Error())
		}
	}

	// Update only those data which are provided in request
	if emp.Name != "" {
		fetchedData.Name = emp.Name
	}
	if emp.Department != "" {
		fetchedData.Department = emp.Department
	}
	if emp.Salary != 0 {
		fetchedData.Salary = emp.Salary
	}

	//update the employee data
	_, conn_err := conn.DbConnect.Exec("UPDATE employee_details SET name=?, department=?, salary=? WHERE id=?", fetchedData.Name, fetchedData.Department, fetchedData.Salary, fetchedData.ID)
	if conn_err != nil {
		return utils.Employee{}, errors.New("Error in Mysql update query :" + conn_err.Error())
	}
	return fetchedData, nil
}
