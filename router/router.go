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
	r.POST("v1/login", middleware.Login)
	r.POST("v1/refresh", middleware.Refresh)

	//Middleware for Rate Limiter
	r.Use(middleware.NewRateLimiter(rate.Limit(100), 10))

	//Middleware for JWT Authentication
	r.Use(middleware.AuthMiddleware)

	//ROUTES--------------
	// imports record from excel file and displays to user
	r.GET("v1/read-excel-sheet", excel.AccessFile)

	// store the Excel records in Mysql and Redis
	r.GET("v1/store-imported-data", excel.StoreImporteddata)

	// Fetches record of specific  employee Id
	r.GET("v1/get_employee/:id", services.GetSingleEmployee)

	// Middleware for RBAC authorization for Super Admin and Admin
	r.POST("v1/create_employee", middleware.RBACMiddleware(middleware.Role_Super_Admin, middleware.Role_Admin), services.CreateEmployee)   //Creates employee record, stores the record in Mysql and Redis
	r.PUT("v1/update_employee", middleware.RBACMiddleware(middleware.Role_Super_Admin, middleware.Role_Admin), services.UpdateEmployeeNew) // Updates record in Redis and then updates redis record to Mysql

	// Middleware for RBAC authorization for Super Admin only
	r.DELETE("v1/delete_employee/:id", middleware.RBACMiddleware(middleware.Role_Super_Admin), services.DeleteEmployee) //Deletes a record
	r.GET("v1/clear_cache", middleware.RBACMiddleware(middleware.Role_Super_Admin), redisops.ClearCache)                //Clear all the caches of Redis
	r.GET("v1/pull-pubsubdata", middleware.RBACMiddleware(middleware.Role_Super_Admin), pubsubops.Pull_PubsubData)      //Pulls all the records present in the pubsub topic

	return r

}
