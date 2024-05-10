package router

import (
	"managedata/excel"

	"managedata/middleware"
	pubsubops "managedata/pubsub_ops"
	redisops "managedata/redisOps"
	"managedata/services"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func NewRouter() *gin.Engine {
	r := gin.New()

	//Middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	//Checks the aliveness of the application without JWT
	r.GET("/", services.Appstart)

	//Generate JWT token using valid credentials
	r.POST("/login", middleware.Login)

	//Middleware for Rate Limiter
	r.Use(middleware.NewRateLimiter(rate.Limit(100), 10))

	//Middleware for JWT Authentication
	r.Use(middleware.AuthMiddleware)

	//ROUTES--------------
	// imports record from excel file and displays to user
	r.GET("/read-excel-sheet", excel.AccessFile)

	// store the Excel records in Mysql and Redis
	r.GET("/store-imported-data", excel.StoreImporteddata)

	// Fetches record of specific  employee Id
	r.GET("/get_employee/:id", services.GetSingleEmployee)

	//Middleware for RBAC authorization
	r.Use(middleware.RBACMiddleware)

	//Creates employee record, stores the record in Mysql and Redis
	r.POST("/create_employee", services.CreateEmployee)

	// Updates record in Redis and then updates redis record to Mysql
	r.PUT("/update_employee", services.UpdateEmployeeNew)

	//Deletes a record
	r.DELETE("/delete_employee/:id", services.DeleteEmployee)

	//Clear all the caches of Redis
	r.GET("/clear_cache", redisops.ClearCache)

	//Pulls all the records present in the pubsub topic
	r.GET("/pull-pubsubdata", pubsubops.Pull_PubsubData)

	return r

}
