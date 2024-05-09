package main

import (
	"fmt"
	"log"
	excel "managedata/excel"
	initpack "managedata/init_pack"
	"managedata/middleware"
	pubsubops "managedata/pubsub_ops"
	"managedata/router"

	"net/http"
)

func main() {

	//CRUD operation routers
	r := router.CrudOperation()

	// imports record from excel file and displays to user
	r.Handle("/read-excel-sheet", middleware.AuthMiddleware(http.HandlerFunc(excel.AccessFile)))

	// store the Excel records in Mysql and Redis
	r.Handle("/store-imported-data", middleware.AuthMiddleware(http.HandlerFunc(excel.StoreImporteddata)))

	r.Handle("/pull-pubsubdata", middleware.AuthMiddleware(http.HandlerFunc(pubsubops.Pull_PubsubData)))

	// Start the HTTP server
	fmt.Println("Application ready to listen: ")
	log.Fatal(http.ListenAndServe(":8080", r))

	// stop the Pubsub topic
	defer initpack.Topic.Stop()

}
