package db

import (
	"fmt"
	"log"
)

func CreateTableIfNotExists(tablename string, column []string) error {
	createquery := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		%s INT AUTO_INCREMENT PRIMARY KEY,
		%s VARCHAR(255) NOT NULL,
		%s VARCHAR(255),
		%s DECIMAL(10, 2)
	);
`, tablename, column[0], column[1], column[2], column[3])
	// Create the employee_details table if it doesn't exist
	_, err := Conn.Exec(createquery)
	if err != nil {
		fmt.Println("table already exist")
		log.Fatal(err)
	} else {
		fmt.Println("Table created successfully query ")
	}
	return err
}
