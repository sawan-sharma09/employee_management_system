package redisops

import (
	"errors"
	"fmt"
	"log"
	initpack "managedata/init_pack"
	"managedata/util"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gomodule/redigo/redis"
)

// Store the data in Redis as HashMap
func RedisSet(emp util.Employee) {
	redisKey := fmt.Sprintf("employee:%d", emp.ID)
	redisconn := initpack.RedisPool.Get()
	_, err := redisconn.Do("HMSET", redisKey, "id", emp.ID, "name", emp.Name, "department", emp.Department, "salary", emp.Salary)
	if err != nil {
		fmt.Println("error in redis set operation :", err) // log.println is not used to avoid stoppage of program for Cache set issue
		return
	} else {
		fmt.Println("Data set successfully in Redis :", emp)
	}
	redisconn.Close()
}

func RedisGet(emp util.Employee) util.Employee {
	// first fetch all the existing data from redis and store in fetchedData, then replace the new data from emp to fetchedData

	// to store fetched data
	var fetchedData util.Employee
	redisKey := fmt.Sprintf("employee:%d", emp.ID)
	red_conn := initpack.RedisPool.Get()
	datadb, err := redis.Values(red_conn.Do("HMGET", redisKey, "id", "name", "department", "salary"))
	if err != nil {
		fmt.Println("error in redis hgetall: ", err)
		red_conn.Close()
	} else {
		_, scan_err := redis.Scan(datadb, &fetchedData.ID, &fetchedData.Name, &fetchedData.Department, &fetchedData.Salary)
		defer red_conn.Close()
		if scan_err != nil {
			fmt.Println("error in redis scan : ", scan_err)
		} else {
			// store both unchanged value and updated record in fetchedData
			if emp.Name != "" {
				fetchedData.Name = emp.Name
			}
			if emp.Department != "" {
				fetchedData.Department = emp.Department
			}
			if emp.Salary != 0 {
				fetchedData.Salary = emp.Salary
			}

		}
	}
	return fetchedData
}

func RedisDel(key string) error {
	conn := initpack.RedisPool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}
	return nil
}

func RedisKeyExists(redisKey string) error {
	red_conn := initpack.RedisPool.Get()

	exists, redisErr := redis.Bool(red_conn.Do("EXISTS", redisKey))
	if redisErr != nil {
		red_conn.Close()
		return errors.New("error in redis query" + redisErr.Error())
	} else if !exists {
		return errors.New("employee id not found")
	}
	return nil
}

func ClearCache(c *fiber.Ctx) error {
	redisconn := initpack.RedisPool.Get()
	_, err := redisconn.Do("FLUSHALL")
	redisconn.Close()
	if err != nil {
		fmt.Println("Error in Executing RedisCache clear Command ", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Redis Flushall Query error: " + err.Error())
	} else {
		log.Println("Redis Cache Cleared at: ", time.Now())

		// Set the response status and send a success message
		return c.Status(fiber.StatusOK).SendString("Cache Cleared successfully")
	}
}
