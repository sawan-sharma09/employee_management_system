package excel

import (
	"fmt"
	"managedata/app_errors"
	"managedata/db"
	initpack "managedata/init_pack"
	redisops "managedata/redisOps"
	"managedata/util"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func StoreImporteddata(c *gin.Context) {
	rows, err := OpenExcelFile()
	if err != nil {
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrExcelOpen, Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
		c.JSON(http.StatusInternalServerError, logDetails)
		return
	}

	// use fileName to name the Mysql Table
	fileName := getFileName(FilePath)
	fmt.Println(rows[0])
	if table_err := db.CreateTableIfNotExists(fileName, rows[0]); table_err != nil {
		fmt.Println("error in CreateTable:", table_err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: table_err.Error(), Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
		c.JSON(http.StatusInternalServerError, logDetails)
		return
	}

	var emp util.Employee
	var errorStrings []string

	for _, row := range rows[1:] {
		// Read data from Excel row
		if len(row) >= 1 {
			emp.ID, _ = strconv.Atoi(row[0])
			emp.Name = row[1]
			emp.Department = row[2]
			emp.Salary, _ = strconv.ParseFloat(row[3], 64)

			// Insert data into the database
			insertquery := fmt.Sprintf("INSERT INTO %s (%s, %s, %s, %s) VALUES (?, ?, ?, ?)", fileName, rows[0][0], rows[0][1], rows[0][2], rows[0][3])
			_, Conn_err := initpack.DbConn.Exec(insertquery, emp.ID, emp.Name, emp.Department, emp.Salary)
			if Conn_err != nil {
				fmt.Println("Error inserting data into database:", Conn_err.Error())
				errorStrings = append(errorStrings, Conn_err.Error())
			} else {
				fmt.Println("Data inserted Successfully in Mysql Db :", emp)

				// Set the data into Redis
				redisops.RedisSet(emp)
			}
		}
	}
	if len(errorStrings) > 1 {
		logDetails := map[string]interface{}{"Timestamp": time.Now(), "Level": "ERROR", "Message": "An error occurred while inserting data into the database", "Endpoint": c.Request.URL.Path, "Errors": errorStrings, "Status_code": http.StatusInternalServerError}
		c.JSON(http.StatusInternalServerError, logDetails)
		return
	}
	c.String(http.StatusOK, "Data imported successfully")

}
