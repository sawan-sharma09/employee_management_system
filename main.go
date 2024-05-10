package main

import (
	"fmt"
	initpack "managedata/init_pack"
	"managedata/router"
)

func main() {

	r := router.NewRouter()
	serverErr := r.Listen(":8080")
	if serverErr != nil {
		fmt.Println("Failed to listen & serve due to error : ", serverErr)
	}

	// stop the Pubsub topic
	defer initpack.Topic.Stop()

}
