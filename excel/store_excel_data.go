package excel

import (
	"fmt"
	"log"
	"managedata/db"
	initpack "managedata/init_pack"
	redisops "managedata/redisOps"
	"managedata/util"

	"net/http"
	"strconv"
)

func StoreImporteddata(w http.ResponseWriter, r *http.Request) {
	rows, err := OpenExcelFile()
	if err != nil {
		log.Println("Error in Opening Excel file")
	}

	// use fileName to name the Mysql Table
	fileName := getFileName(FilePath)
	fmt.Println(rows[0])
	if table_err := db.CreateTableIfNotExists(fileName, rows[0]); table_err != nil {
		fmt.Println("error in CreateTable:", table_err)
		http.Error(w, table_err.Error(), http.StatusInternalServerError)
		return
	}

	var emp util.Employee
	for _, row := range rows[1:] {
		// Read data from Excel row
		if len(row) >= 1 {
			emp.ID, _ = strconv.Atoi(row[0])
			emp.Name = row[1]
			emp.Department = row[2]
			emp.Salary, _ = strconv.ParseFloat(row[3], 64)

			// Insert data into the database
			insertquery := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)", fileName, rows[0][0], rows[0][1], rows[0][2], rows[0][3])
			_, Conn_err := initpack.Conn.Exec(insertquery, emp.ID, emp.Name, emp.Department, emp.Salary)
			if Conn_err != nil {
				log.Println("Error inserting data:", Conn_err)
			} else {
				fmt.Println("Data inserted Successfully in Mysql Db :", emp)

				// Set the data into Redis
				redisops.RedisSet(emp)
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Data imported successfully\n")
}
