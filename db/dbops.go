package db

import (
	"errors"
	"fmt"
	initpack "managedata/init_pack"
)

func CreateTableIfNotExists(tablename string, column []string) error {
	tableExists, err := TableExists(tablename)
	if err != nil {
		return err
	}
	if tableExists {
		return errors.New("table already exists")
	}

	createQuery := fmt.Sprintf(`
	CREATE TABLE %s (
		%s INT AUTO_INCREMENT PRIMARY KEY,
		%s VARCHAR(255) NOT NULL,
		%s VARCHAR(255),
		%s DECIMAL(10, 2)
	);
`, tablename, column[0], column[1], column[2], column[3])

	_, err = initpack.DbConn.Exec(createQuery)
	if err != nil {
		return err
	}

	fmt.Println("Table created successfully:", tablename)
	return nil
}

func TableExists(tableName string) (bool, error) {
	query := fmt.Sprintf("SHOW TABLES LIKE '%s'", tableName)
	rows, err := initpack.DbConn.Query(query)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	return rows.Next(), nil
}
