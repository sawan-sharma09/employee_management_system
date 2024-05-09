package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	grpcServices "managedata/grpc_services/grpc_client"
	initpack "managedata/init_pack"

	redisops "managedata/redisOps"
	"managedata/util"
	"time"

	"net/http"

	"github.com/gorilla/mux"
)

func GetSingleEmployee(w http.ResponseWriter, r *http.Request) {

	if !initpack.Limiter.Allow() {
		http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		return
	}

	// Retrieve an employee by ID from the database
	params := mux.Vars(r)
	id := params["id"]

	var emp util.Employee
	err := initpack.Conn.QueryRow("SELECT * FROM employee_details WHERE id=?", id).Scan(&emp.ID, &emp.Name, &emp.Department, &emp.Salary)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Employee not found", http.StatusNotFound)
			return
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println("error in Mysql Select operation", err)
		}
		return
	}

	fmt.Printf("Fetched employee details: %+v\n", emp)

	// Convert the result to JSON and write to the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(emp)
}

func CreateEmployee(w http.ResponseWriter, r *http.Request) {

	if !initpack.Limiter.Allow() {
		http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		return
	}

	var emp util.Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		fmt.Println("error in NewDecoder: ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Insert the new employee into the Mysql database
	_, conn_err := initpack.Conn.Exec("INSERT INTO employee_details (id,name, department, salary) VALUES (? ,?, ?, ?)", emp.ID, emp.Name, emp.Department, emp.Salary)
	if conn_err != nil {
		fmt.Println("error in Mysql insert operation", conn_err)
		// w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, "Mysql Insert Error: "+conn_err.Error(), http.StatusInternalServerError)
		return
	} else {
		fmt.Println("Employee created Successfully :", emp)
	}

	//Set data into redis
	redisops.RedisSet(emp)

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Employee created with ID: %d\n", emp.ID)
}

func DeleteEmployee(w http.ResponseWriter, r *http.Request) {

	if !initpack.Limiter.Allow() {
		http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		return
	}

	params := mux.Vars(r)
	id := params["id"]

	redisKey := fmt.Sprintf("employee:%s", id)

	//Check if the key for the Employee id exists or not
	if keyErr := redisops.RedisKeyExists(redisKey); keyErr != nil {
		fmt.Println(keyErr)
		return
	}

	// Delete the employee data from the  Mysql database
	_, deleteErr := initpack.Conn.Exec("DELETE FROM employee_details WHERE id=?", id)
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

func UpdateEmployee(w http.ResponseWriter, r *http.Request) {

	if !initpack.Limiter.Allow() {
		fmt.Println("Rate limit Exceeded")
		http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		return
	}

	var emp util.Employee
	err := json.NewDecoder(r.Body).Decode(&emp)
	if err != nil {
		fmt.Println("Error in NewDecoder of UpdateEmployee", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}


	streamErr := grpcServices.Bidi_stream(emp)
	if streamErr != nil {
		fmt.Println("Error in Grpc Service: ", streamErr)
		return
	} else {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Employee updated successfully\n")
	}

	//Update the data in Redis Cache
	// redisops.RedisSet(fetchedData)

	// // Set the response status and send a success message
	// w.WriteHeader(http.StatusOK)
	// fmt.Printf("Employee updated \nData: %v\n", fetchedData)

	//Publish the updated data to Pubsub topic
	// id, publishErr := pubsubops.PublishData(fetchedData)
	// if publishErr != nil {
	// 	fmt.Println("Error in Pubsub Publish: ", publishErr)
	// } else {
	// 	fmt.Fprintf(w, "Published message with msg ID: %v\n", id)
	// }

}

func ClearCache(w http.ResponseWriter, r *http.Request) {

	if !initpack.Limiter.Allow() {
		fmt.Println("Rate limit Exceeded")
		http.Error(w, "Rate Limit Exceeded", http.StatusTooManyRequests)
		return
	}

	redisconn := initpack.RedisPool.Get()
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

// gRPC Bidirectional call
// func GrpcUpdate(w http.ResponseWriter, r *http.Request) {
// 	grpcServices.Grpc_Call()
// }
