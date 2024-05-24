package initpack

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	DbConn    *sql.DB
	RedisPool *redis.Pool
	Client    *pubsub.Client
	Topic     *pubsub.Topic
	Cc        *grpc.ClientConn
)

func init() {
	InitConn()
}
func InitConn() {

	var err error
	var mysqlURL string

	enverr := godotenv.Load("./config/secret.env")
	if enverr != nil {
		log.Fatal("Error loading .env file", enverr)
	}

	env := flag.String("env", "", "Specify the environment(dev/staging)")
	flag.Parse()

	//env flag
	switch *env {
	case "dev":
		fmt.Println("Running in dev environment")
		mysqlURL = os.ExpandEnv("$MYSQL_DB_URL")
	case "stage":
		fmt.Println("Running in stage environment")
	default:
		log.Fatal("Invalid environment. Please specify 'dev' or 'stage'.")
	}

	// Open a database connection
	DbConn, err = sql.Open("mysql", mysqlURL)
	if err != nil {
		log.Fatal(err)
	}

	DbConn.SetMaxOpenConns(100)                //The maximum number of open connections in the pool. This option controls the concurrency of database access. Once this limit is reached, further requests for connections will block until a connection becomes available.
	DbConn.SetMaxIdleConns(30)                 // Maximum number of connections that can remain idle (i.e., not in use) in the pool at any given time. Keeping a certain number of idle connections can help improve performance by reducing the overhead of establishing new connections for subsequent database operations.
	DbConn.SetConnMaxLifetime(0)               // Maximum lifetime of a connection (0 means no limit)
	DbConn.SetConnMaxIdleTime(2 * time.Minute) // Maximum time a connection can be idle before it's closed

	// Check if the connection is successful
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if pingErr := DbConn.PingContext(dbCtx); pingErr != nil {
		log.Fatal("Error pinging MySQL database: ", pingErr)
	} else {
		fmt.Println("MySQL db connected successfully")
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

	//gRPC Connection
	var grpcErr error
	Cc, grpcErr = grpc.Dial("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if grpcErr != nil {
		fmt.Println("Error in connecting to client: ", err)
	}

}
