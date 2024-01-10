package main

import (
	"fmt"
	"log"
	excel "managedata/excel"
	"managedata/router"

	"net/http"
)

func main() {

	//CRUD operation routers
	r := router.CrudOperation()

	// imports record from excel file and displays to user
	r.HandleFunc("/read-excel-sheet", excel.AccessFile)

	// store the Excel records in Mysql and Redis
	r.HandleFunc("/store-imported-data", excel.StoreImporteddata)

	// Start the HTTP server
	fmt.Println("Application ready to listen: ")
	log.Fatal(http.ListenAndServe(":7070", r))
}
