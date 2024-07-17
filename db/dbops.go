package db

import (
	"context"
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
		%s SERIAL PRIMARY KEY,
		%s VARCHAR(255) NOT NULL,
		%s VARCHAR(255),
		%s DECIMAL(10, 2)
	);
`, tablename, column[0], column[1], column[2], column[3])

	_, execErr := initpack.PostgresPool.Exec(context.Background(), createQuery)
	if execErr != nil {
		return err
	}

	fmt.Println("Table created successfully:", tablename)
	return nil
}

func TableExists(tableName string) (bool, error) {

	query := `
	SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = $1
	);
`
	var exists bool
	err := initpack.PostgresPool.QueryRow(context.Background(), query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
