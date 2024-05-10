package excel

import (
	"fmt"
	"managedata/db"
	initpack "managedata/init_pack"
	redisops "managedata/redisOps"
	"managedata/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func StoreImporteddata(c *gin.Context) {
	rows, err := OpenExcelFile()
	if err != nil {
		fmt.Println("Error in Opening Excel file")
		c.String(http.StatusInternalServerError, "Excel file opening error:"+err.Error())
		return
	}

	// use fileName to name the Mysql Table
	fileName := getFileName(FilePath)
	fmt.Println(rows[0])
	if table_err := db.CreateTableIfNotExists(fileName, rows[0]); table_err != nil {
		fmt.Println("error in CreateTable:", table_err)
		c.String(http.StatusInternalServerError, "Error creating table --> "+table_err.Error())
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
				fmt.Println("Error inserting data:", Conn_err)
				c.String(http.StatusInternalServerError, "Error inserting data: "+Conn_err.Error())
			} else {
				fmt.Println("Data inserted Successfully in Mysql Db :", emp)

				// Set the data into Redis
				redisops.RedisSet(emp)
			}
		}
	}
	c.String(http.StatusOK, "Data imported successfully")
}
