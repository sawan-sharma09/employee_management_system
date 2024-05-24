package excel

import (
	"fmt"
	"managedata/app_errors"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

var FilePath string = "./docs/employee_details.xlsx"

func OpenExcelFile() ([][]string, error) {
	xlsx, err := excelize.OpenFile(FilePath)
	if err != nil {
		fmt.Println("Error in opening excel file :", err)
		return nil, err
	}

	// Assuming data is present in the first sheet (index 1).
	sheetName := xlsx.GetSheetName(0)
	// 2d slice for storing data
	rows, row_err := xlsx.GetRows(sheetName)
	if row_err != nil {
		fmt.Println("Error in get rows: ", row_err)
		return nil, row_err
	} else {
		return rows, nil
	}
}
func readExcelFile() ([]map[string]string, error) {

	// Open the excel file and return fetched rows
	rows, err := OpenExcelFile()
	if err != nil {
		return nil, err
	}

	// Assuming the first row contains column headers.
	columnHeaders := rows[0]

	var data []map[string]string

	for _, row := range rows[1:] {
		entry := make(map[string]string)

		for i, cellValue := range row {
			if i < len(columnHeaders) {
				fmt.Printf("%s: %s\t", columnHeaders[i], cellValue)
				entry[columnHeaders[i]] = cellValue
			}
		}
		// all the records fetched from excel sheet are appended and stored in "data"
		data = append(data, entry)
		fmt.Println()
	}
	return data, nil
}

func AccessFile(c *gin.Context) {
	data, err := readExcelFile()
	if err != nil {
		fmt.Println("Error reading excel file: ", err)
		logDetails := app_errors.ErrorTemplate{Timestamp: time.Now(), Level: "ERROR", Message: app_errors.ErrFileAccess, Endpoint: c.Request.URL.Path, Status_code: http.StatusInternalServerError}
		c.AbortWithStatusJSON(http.StatusInternalServerError, logDetails)
		return
	}
	c.JSON(http.StatusOK, &data)
}

func getFileName(filePath string) string {
	// Use filepath.Base to get the base name of the file
	baseName := filepath.Base(filePath)

	// Remove the file extension if present
	fileName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	return fileName
}
