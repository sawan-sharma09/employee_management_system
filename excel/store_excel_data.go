package excel

import (
	"fmt"
	"managedata/db"
	initpack "managedata/init_pack"
	redisops "managedata/redisOps"
	"managedata/util"
	"strconv"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// StoreImporteddata stores data imported from an Excel file into a MySQL database
func StoreImporteddata(c *fiber.Ctx) error {
	rows, err := OpenExcelFile()
	if err != nil {
		fmt.Println("Error in Opening Excel file")
		return c.Status(fiber.StatusInternalServerError).SendString("Excel file opening error: ")
	}

	// use fileName to name the Mysql Table
	fileName := getFileName(FilePath)
	fmt.Println(rows[0])

	// Create MySQL table if it does not exist
	if table_err := db.CreateTableIfNotExists(fileName, rows[0]); table_err != nil {
		fmt.Println("error in CreateTable:", table_err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating table --> " + table_err.Error())
	}

	// Define variables for concurrency and error handling
	var (
		emp    util.Employee
		wg     sync.WaitGroup
		mu     sync.Mutex
		errors []error
	)

	// Insert data into the database concurrently
	for _, row := range rows[1:] {
		wg.Add(1)

		// Read data from the Excel row and insert into the database
		go func(row []string) {
			defer wg.Done()
			if len(row) >= 1 {
				emp.ID, _ = strconv.Atoi(row[0])
				emp.Name = row[1]
				emp.Department = row[2]
				emp.Salary, _ = strconv.ParseFloat(row[3], 64)

				// Insert data into the database
				insertquery := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)", fileName, rows[0][0], rows[0][1], rows[0][2], rows[0][3])
				_, Conn_err := initpack.Conn.Exec(insertquery, emp.ID, emp.Name, emp.Department, emp.Salary)
				if Conn_err != nil {
					fmt.Println("Error inserting data:", Conn_err)

					//append all the errors to send a slice of error to client
					mu.Lock()
					errors = append(errors, Conn_err)
					mu.Unlock()

				} else {
					fmt.Println("Data inserted Successfully in Mysql Db :", emp)

					// Set the data into Redis
					redisops.RedisSet(emp)
				}
			}
		}(row)
	}

	wg.Wait()

	// If errors occurred during data insertion, send them as a JSON response
	// The errors are sent in a slice because Fiber doesn't support sending multiple error responses in a loop.
	if len(errors) > 1 {
		return c.Status(fiber.StatusInternalServerError).JSON(errors)
	}

	// If no error occurred during data insertion, send success response
	return c.Status(fiber.StatusOK).SendString("Data imported successfully")
}
