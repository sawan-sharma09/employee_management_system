package initpack

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	PostgresPool *pgxpool.Pool
	RedisPool    *redis.Pool
	Client       *pubsub.Client
	Topic        *pubsub.Topic
	Cc           *grpc.ClientConn
)

func init() {
	InitConn()
}
func InitConn() {

	var err error
	var redisURL, postgresURL string

	enverr := godotenv.Load("./config/secret.env")
	if enverr != nil {
		fmt.Println("Error loading .env file for local", enverr)
	}

	//env for dockerfile--ignore the error message in local system
	docErr := godotenv.Load("/app/secret.env")
	if docErr != nil {
		fmt.Println("Error loading .env file for docker image")
	}

	env := flag.String("env", "", "Specify the environment(dev/staging/docker)")
	flag.Parse()

	//env flag
	switch *env {
	case "dev":
		fmt.Println("Running in dev environment")
		redisURL = os.ExpandEnv("$REDIS_URL")
		postgresURL = os.ExpandEnv("$POSTGRES_DB")
	case "stage":
		fmt.Println("Running in stage environment")
	case "docker":
		fmt.Println("Running in docker container environment")
		redisURL = os.ExpandEnv("$REDIS_URL_DOCKER")
	default:
		log.Fatal("Invalid environment. Please specify 'dev' or 'stage'.")
	}

	//postgres connection
	// config, err := pgxpool.ParseConfig(postgresURL + "?sslmode=disable")
	config, err := pgxpool.ParseConfig(postgresURL)

	if err != nil {
		log.Fatal("Error configuring the postgres database CSC: ", err)
	}
	config.MaxConns = 100                               //The maximum number of open connections in the pool. This option controls the concurrency of database access. Once this limit is reached, further requests for connections will block until a connection becomes available.
	config.MinConns = 30                                //The minimum number of connections to keep open in the pool. Connections below this threshold will be opened to meet this requirement.
	config.MaxConnLifetime = 0                          // Maximum lifetime of a connection (0 means no limit)
	config.MaxConnIdleTime = 10 * time.Minute           // Maximum time a connection can be idle before it's closed
	config.HealthCheckPeriod = 5 * time.Second          // Frequency of health checks (0 means no health checks)
	config.ConnConfig.ConnectTimeout = 10 * time.Second // Maximum time to establish a new connection

	var configErr error
	PostgresPool, configErr = pgxpool.NewWithConfig(context.Background(), config)
	if configErr != nil {
		log.Fatal("Error connecting to the postgres database CSC: ", configErr)
	} else {
		fmt.Println("Postgres Db Connected Successfully for CSC")
	}

	//redis
	RedisPool = &redis.Pool{
		MaxIdle:     100,
		MaxActive:   200,
		IdleTimeout: 10 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisURL, redis.DialTLSSkipVerify(true), redis.DialConnectTimeout(time.Duration(2)*time.Second))
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
