package router

import (
	"managedata/middleware"
	"managedata/services"
	"net/http"

	"github.com/gorilla/mux"
)

func CrudOperation() *mux.Router {
	r := mux.NewRouter()

	//jwt-> authentication
	r.HandleFunc("/login", middleware.Login)

	secure := r.PathPrefix("/secure").Subrouter()
	secure.Use(middleware.AuthMiddleware)
	secure.HandleFunc("/get_employee/{id}", services.GetSingleEmployee).Methods("GET")

	//Creates employee record, stores the record in Mysql and Redis
	secure.Handle("/create_employee", middleware.RBACMiddleware(http.HandlerFunc(services.CreateEmployee))).Methods("POST")

	// Updates record in Redis and then updates redis record to Mysql
	secure.Handle("/update_employee", middleware.RBACMiddleware(http.HandlerFunc(services.UpdateEmployee))).Methods("PUT")

	//Deletes a record
	secure.Handle("/delete_employee/{id}", middleware.RBACMiddleware(http.HandlerFunc(services.DeleteEmployee))).Methods("DELETE")

	//cache clear
	secure.Handle("/clear_cache", middleware.RBACMiddleware(http.HandlerFunc(services.ClearCache)))

	return r
}
