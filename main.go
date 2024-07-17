package main

import (
	"fmt"
	initpack "managedata/init_pack"
	"managedata/router"
)

//	@title			EMS API
//	@version		1.0
//	@description	This is an Employee Management System API

//	@host		localhost:8080
//	@BasePath	/
//	@schemes	http https

// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				"JWT Authorization header using the Bearer scheme. Example: \"Authorization: Bearer {token}\""
func main() {

	r := router.NewRouter()
	serverErr := r.Run(":8080")
	if serverErr != nil {
		fmt.Println("Failed to listen & serve due to error : ", serverErr)
	}

	// stop the Pubsub topic
	defer initpack.Topic.Stop()
	defer initpack.PostgresPool.Close()

}
