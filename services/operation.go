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

	var emp util.Employee
	err := db.Conn.QueryRow("SELECT * FROM employee_details WHERE id=?", id).Scan(&emp.ID, &emp.Name, &emp.Department, &emp.Salary)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Employee not found", http.StatusNotFound)
		} else {
			log.Fatal("error in Mysql Select operation", err)
		}
		return
	}

	fmt.Printf("Fetched employee details: %+v", emp)

	// Convert the result to JSON and write to the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(emp)
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

// func RedisKeyExists(redisKey string) error {
// 	red_conn := db.RedisPool.Get()

// 	exists, redisErr := redis.Bool(red_conn.Do("EXISTS", redisKey))

// 	if redisErr != nil {
// 		red_conn.Close()
// 		fmt.Println("RedisErr: ", redisErr)
// 		return errors.New("error in redis query")
// 	} else if !exists {
// 		return errors.New("employee id not found")
// 	}
// 	return nil
// }

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	id := params["id"]

	redisKey := fmt.Sprintf("employee:%s", id)

	//Check if the key for the Employee id exists or not
	if keyErr := redisops.RedisKeyExists(redisKey); keyErr != nil {
		fmt.Println(keyErr)
		return
	}

	// Delete the employee data from the  Mysql database
	_, deleteErr := db.Conn.Exec("DELETE FROM employee_details WHERE id=?", id)
	if deleteErr != nil {
		fmt.Println("err in Mysql delete operation : ", deleteErr)
		log.Fatal(deleteErr)
	}

	// Delete the employee data from redis
	redisDelErr := redisops.RedisDel(redisKey)
	if redisDelErr != nil {
		fmt.Println("err in Redis delete operation: ", redisDelErr)
		log.Fatal(redisDelErr)
	}

	fmt.Printf("EmployeeId: %v deleted\n ", id)
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

	// Retrieve the existing record for employee 'emp' from Redis and store it with updated data in the fetchedData struct.
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
