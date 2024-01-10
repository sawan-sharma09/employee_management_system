package router

import (
	"managedata/services"

	"github.com/gorilla/mux"
)

func CrudOperation() *mux.Router {
	r := mux.NewRouter()

	// Fetches record of specific  employee Id
	r.HandleFunc("/get_employee/{id}", services.GetSingleEmployee).Methods("GET")

	//Creates employee record, stores the record in Mysql and Redis
	r.HandleFunc("/create_employee", services.CreateEmployee).Methods("POST")

	// Updates record in Redis and then updates redis record to Mysql
	r.HandleFunc("/update_employee", services.UpdateEmployeeNew).Methods("PUT")

	//Deletes a record
	r.HandleFunc("/delete_employee/{id}", services.DeleteEmployee).Methods("DELETE")

	//cache clear
	r.HandleFunc("/clear_cache", services.ClearCache)
	return r
}
