package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
)

var (
	Conn      *sql.DB
	RedisPool *redis.Pool
)

func init() {
	InitConn()
}
func InitConn() {
	const envFilePath string = "./config/secret.env"
	enverr := godotenv.Load(envFilePath)
	if enverr != nil {
		log.Fatal("Error loading .env file", enverr)
	}
	var err error
	// Open a database connection
	Conn, err = sql.Open("mysql", os.ExpandEnv("$MYSQL_DB_URL"))
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()

	// Check if the connection is successful
	err = Conn.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connected to MySQL!")
	}
	//redis
	RedisPool = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   200,
		IdleTimeout: 10 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", os.ExpandEnv("$REDIS_URL"), redis.DialTLSSkipVerify(true), redis.DialConnectTimeout(time.Duration(2)*time.Second))
			if err != nil {
				fmt.Println("redis init dial err :", err)
				return nil, err
			}
			return c, err
		},
	}

	// Ping the Redis server to check the connection
	redisconn := RedisPool.Get()
	if _, err := redisconn.Do("PING"); err != nil {
		log.Fatal(err)
		redisconn.Close()
	} else {
		fmt.Println("Connected to Redis")
		redisconn.Close()
	}
}
