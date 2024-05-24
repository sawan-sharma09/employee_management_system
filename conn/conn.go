package conn

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Lis       net.Listener
	DbConnect *sql.DB
)

func init() {
	var err error
	Lis, err = net.Listen("tcp", ":8090")
	if err != nil {
		fmt.Println("Error in listening to port: ", err)
	}

	// envErr := godotenv.Load("./config/secret.env")
	// if envErr != nil {
	// 	fmt.Println("Error loading env file: ", envErr)
	// }

	// Open a database connection
	DbConnect, err = sql.Open("mysql", os.ExpandEnv("root:tiger@tcp(127.0.0.1:3306)/testdb"))
	if err != nil {
		log.Fatal("Sql db connection error ", err)
	}

	// Check if the connection is successful
	err = DbConnect.Ping()
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Connected to MySQL!")
	}

}
