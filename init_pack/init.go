package initpack

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Conn      *sql.DB
	RedisPool *redis.Pool
	Client    *pubsub.Client
	Topic     *pubsub.Topic
	Limiter   *rate.Limiter
	Cc        *grpc.ClientConn
)

func init() {
	InitConn()
}
func InitConn() {

	enverr := godotenv.Load("./config/secret.env")
	if enverr != nil {
		log.Fatal("Error loading .env file", enverr)
	}

	var err error

	// Open a database connection
	Conn, err = sql.Open("mysql", os.ExpandEnv("$MYSQL_DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

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

	//Pubsub connection
	projectID := os.Getenv("PROJECTID")
	ctx := context.Background()

	var clientErr error
	Client, clientErr = pubsub.NewClient(ctx, projectID, option.WithCredentialsFile(os.Getenv("PUBSUB_CREDENTIALS")))
	if err != nil {
		log.Fatal("Error in creating client..", clientErr)
	}

	Topic = Client.Topic("TestTopic")

	// Define the rate limiter configuration with a rate limit of 100 requests per second and a burst size of 10.
	Limiter = rate.NewLimiter(rate.Limit(100), 10)

	//gRPC Connection
	var grpcErr error
	Cc, grpcErr = grpc.Dial("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if grpcErr != nil {
		fmt.Println("Error in connecting to client: ", err)
	}

}
