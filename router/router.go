package router

import (
	"managedata/excel"
	"managedata/middleware"
	pubsubops "managedata/pubsub_ops"
	redisops "managedata/redisOps"
	"managedata/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

func NewRouter() *fiber.App {

	r := fiber.New(fiber.Config{
		Prefork:       false,         //true in prod//Prefork:Impact: Setting Prefork to true enables preforking, which can improve performance in a production environment. Each worker process handles incoming requests independently.Recommendation: Use Prefork: true in production for better performance.
		StrictRouting: true,          //Impact: When StrictRouting is set to true, the router will only match routes exactly and not consider trailing slashes. For example, "/user" won't match "/user/".Recommendation: Keep StrictRouting: true if you want strict route matching.
		CaseSensitive: true,          //Impact: When CaseSensitive is set to true, the router considers route paths as case-sensitive.Recommendation: Adjust based on your preference. If you want case-sensitive routing, keep it as true.
		Concurrency:   1024 * 1024,   //Impact: Defines the maximum number of concurrent connections. This sets the maximum number of simultaneous requests your server can handle.Recommendation: Adjust based on your server's capacity. The value 1024 * 1024 allows a large number of concurrent connections. You might want to adjust this based on your server's capacity and expected traffic.
		AppName:       "INT-FINTECH", //Impact: This is just a label for your application. It doesn't affect the server's behavior but can be useful for identification in logs or monitoring tools.Recommendation: Set it to a meaningful name for your application.
		//Impact: These settings control how long the server will wait for various actions before timing out.
		ReadTimeout:  100 * time.Second, //Maximum duration for reading the entire request, including body.
		WriteTimeout: 100 * time.Second, //Maximum duration for writing the response.
		IdleTimeout:  3 * time.Minute,   //Maximum amount of time to wait for the next request when keep-alives are enabled.
	})

	//Checks the aliveness of the application
	r.Get("/", services.Appstart)

	//Generate JWT token using valid credentials
	r.Post("/login", middleware.Login)

	//Middleware for Rate Limiter
	r.Use(middleware.NewRateLimiter(rate.Limit(100), 10))

	//Middleware for JWT Authentication
	r.Use(middleware.AuthMiddleware)

	// imports record from excel file and displays to user
	r.Get("/read-excel-sheet", excel.AccessFile)

	// store the Excel records in Mysql and Redis
	r.Get("/store-imported-data", excel.StoreImporteddata)

	// Fetches record of specific  employee Id
	r.Get("/get_employee/:id", services.GetSingleEmployee)

	//Middleware for RBAC Authorization
	r.Use(middleware.RBACMiddleware)

	//Creates employee record, stores the record in Mysql and Redis
	r.Post("/create_employee", services.CreateEmployee)

	// Updates record in Redis and then updates redis record to Mysql
	r.Put("/update_employee", services.UpdateEmployeeNew)

	//Deletes a record
	r.Delete("/delete_employee/:id", services.DeleteEmployee)

	//Clear all the caches of Redis
	r.Get("/clear_cache", redisops.ClearCache)

	//Pulls all the records present in the pubsub topic
	r.Get("/pull-pubsubdata", pubsubops.Pull_PubsubData)

	return r
}
