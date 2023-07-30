package postgresql

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"
)

func (c *PgClient) CheckTablePresence(tables []interface{}) error {
	for _, table := range tables {
		typeOfTable := reflect.TypeOf(table)
		isPresent, err := c.IsTablePresent(typeOfTable.Name())
		if err != nil {
			fmt.Errorf("failed to check table %v", err)
			return err
		}
		if isPresent {
			// c.Logger.Info(fmt.Sprintf("Table %s exists", typeOfTable.Name()))
			fmt.Sprintf("Table %s exists", err)
		} else {
			err = c.CreateTable(table)
			if err != nil {
				fmt.Errorf("failed to create table %v", err)
				return err
			}
			// c.Logger.Infof("Table %s created", typeOfTable.Name())
			fmt.Sprintf("Table %s created", err)
		}
	}
	return nil
}

func (c *PgClient) IsTablePresent(tableName string) (bool, error) {
	tableNameToLower := strings.ToLower(tableName)
	// Query checks if a table with the given name exists
	query := `
	SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public'
		AND table_name = $1
	);
	`

	var exists bool
	err := c.Pool.QueryRow(context.TODO(), query, tableNameToLower).Scan(&exists)
	if err != nil {
		return false, err
	}

	// If the table does not exist, return false
	return exists, nil
}

func (c *PgClient) CreateTable(table interface{}) error {
	typeOfTable := reflect.TypeOf(table)

	// Get the table name from the type
	tableName := strings.ToLower(typeOfTable.Name())

	// Generate the SQL statement to create the table
	createTableSQL := generateCreateTableSQL(tableName, typeOfTable)

	// Execute the create table query
	_, err := c.Pool.Exec(context.TODO(), createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table %s: %v", tableName, err)
	}

	return nil
}

// Helper function to generate the SQL statement to create a table based on the struct type
func generateCreateTableSQL(tableName string, tableType reflect.Type) string {
	var createTableSQL strings.Builder
	createTableSQL.WriteString("CREATE TABLE IF NOT EXISTS ")
	createTableSQL.WriteString(`"`)
	createTableSQL.WriteString(tableName)
	createTableSQL.WriteString(`" (`)

	numFields := tableType.NumField()
	for i := 0; i < numFields; i++ {
		field := tableType.Field(i)
		columnName := field.Tag.Get("bson")
		columnType := getColumnType(field.Type)

		if columnName == "" {
			continue
		}

		createTableSQL.WriteString(`"`)
		createTableSQL.WriteString(columnName)
		createTableSQL.WriteString(`" `)
		createTableSQL.WriteString(columnType)
		// Add comma if it's not the last column
		if i < numFields-1 {
			createTableSQL.WriteString(",")
		}
	}

	createTableSQL.WriteString(")")

	return createTableSQL.String()
}

// Helper function to map Go types to PostgreSQL column types
func getColumnType(field reflect.Type) string {
	switch field.Kind() {
	case reflect.String:
		return "text"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "real"
	case reflect.Bool:
		return "boolean"
	case reflect.Struct:
		if field == reflect.TypeOf(time.Time{}) {
			return "timestamp"
		}
	case reflect.Slice:
		if field.Elem().Kind() == reflect.Uint8 {
			return "bytea"
		}
	}

	return "text"
}

func (c *PgClient) TableExists(tableName string) (bool, error) {
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = $1)"
	var exists bool
	err := c.Pool.QueryRow(context.TODO(), query, tableName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
