package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"managedata/db"
	redisops "managedata/redisOps"
	"managedata/util"
	"time"

	// redisdb "managedata/redis"
	"net/http"

	"github.com/gorilla/mux"
)

func GetSingleEmployee(w http.ResponseWriter, r *http.Request) {
	// Retrieve an employee by ID from the database
	params := mux.Vars(r)
	id := params["id"]

	var e util.Employee
	err := db.Conn.QueryRow("SELECT * FROM employee_details WHERE id=?", id).Scan(&e.ID, &e.Name, &e.Department, &e.Salary)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Employee not found", http.StatusNotFound)
		} else {
			log.Fatal("error in Mysql Select operation", err)
		}
		return
	}

	// Convert the result to JSON and write to the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func CreateEmployee(w http.ResponseWriter, r *http.Request) {

	var emp util.Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		fmt.Println("error in NewDecoder: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Insert the new employee into the Mysql database
	_, conn_err := db.Conn.Exec("INSERT INTO employee_details (id,name, department, salary) VALUES (? ,?, ?, ?)", emp.ID, emp.Name, emp.Department, emp.Salary)
	if conn_err != nil {
		log.Fatal("error in Mysql insert operation", conn_err)
	} else {
		fmt.Println("Employee created Successfully :", emp)
	}

	//Set data into redis
	redisops.RedisSet(emp)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Employee created with ID: %d\n", emp.ID)
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	id := params["id"]

	// Delete the employee data from the  Mysql database
	_, err := db.Conn.Exec("DELETE FROM employee_details WHERE id=?", id)
	if err != nil {
		fmt.Println("err in Mysql delete operation : ", err)
		log.Fatal(err)
	}

	// Delete the employee data from redis
	redisKey := fmt.Sprintf("employee:%s", id)
	redisErr := redisops.RedisDel(redisKey)
	if redisErr != nil {
		fmt.Println("err in Redis delete operation: ", redisErr)
		log.Fatal(redisErr)
	}
	// Set the response status and send a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Employee deleted successfully\n")
}

func UpdateEmployeeNew(w http.ResponseWriter, r *http.Request) {
	var emp util.Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		fmt.Println("Error in NewDecoder of UpdateEmployee", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Retrieve the existing record for employee 'emp' from Redis and store it  with updated data in the fetchedData struct.
	fetchedData := redisops.RedisGet(emp)

	//Update the data in Redis Cache
	redisops.RedisSet(fetchedData)

	//Update fetched updated data from Redis to Mysql
	_, conn_err := db.Conn.Exec("UPDATE employee_details SET name=?, department=?, salary=? WHERE id=?", fetchedData.Name, fetchedData.Department, fetchedData.Salary, fetchedData.ID)
	if conn_err != nil {
		log.Fatal(" Error in Mysql update query :", conn_err)
	}

	// Set the response status and send a success message
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Employee updated successfully\n")

}

func ClearCache(w http.ResponseWriter, r *http.Request) {
	redisconn := db.RedisPool.Get()
	_, err := redisconn.Do("FLUSHALL")
	redisconn.Close()
	if err != nil {
		log.Println("Error in Executing RedisCache clear Command ", err)
	} else {
		log.Println("Redis Cache Cleared at: ", time.Now())
		// Set the response status and send a success message
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Cache Cleared successfully\n")
	}
}
