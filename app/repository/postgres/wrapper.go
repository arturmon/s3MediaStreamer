package postgres

import (
	"context"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// TransactionFunc defines the signature of the function that will be executed within a transaction.
type TransactionFunc func(ctx context.Context, tx pgx.Tx) error

// ExecuteInTransaction executes a function within a database transaction.
// It handles transaction commit and rollback.
func (c *Client) ExecuteInTransaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := c.Pool.Begin(ctx)
	if err != nil {
		return err
	}

	// Ensure transaction is rolled back if the function returns an error
	defer func() {
		if rErr := tx.Rollback(ctx); rErr != nil && err == nil {
			err = rErr
		}
	}()

	// Execute the provided function
	err = fn(ctx, tx)
	if err != nil {
		return err
	}

	// Commit transaction if function completes successfully
	err = tx.Commit(ctx)
	return err
}

// ExecuteSQL executes a SQL query within the provided transaction.
func ExecuteSQL(ctx context.Context, tx pgx.Tx, query squirrel.Sqlizer) error {
	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, sql, args...)
	return err
}

// TODO:

// QueryMapsWithLogging wraps the QueryMaps function to add custom error handling functionality.
// It converts a Squirrel query into SQL, executes it, and returns the results as a slice of maps.
// This function is useful for executing dynamic SQL queries while providing detailed error context.
// Usage Example:
//
// conn, err := pgx.Connect(context.Background(), "your_connection_string")
//
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// defer conn.Close(context.Background())
//
// query := squirrel.Select("*").From("your_table").Where(squirrel.Eq{"column_name": "value"})
// results, err := QueryMapsWithLogging(conn, query)
//
//	if err != nil {
//	    // Handle the error appropriately
//	    fmt.Println("Error:", err)
//	} else {
//
//	    // Process the results
//	    fmt.Println("Results:", results)
//	}
func QueryMapsWithLogging(conn *pgx.Conn, query squirrel.Sqlizer) ([]map[string]interface{}, error) {
	sql, args, err := query.ToSql() // Convert squirrel query to SQL
	if err != nil {
		return nil, fmt.Errorf("%w: %s", fmt.Errorf("error generating SQL query"), err)
	}

	// Call the original QueryMaps function
	results, err := QueryMaps(conn, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", fmt.Errorf("error executing query"), err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for query: %s", sql)
	}

	return results, nil // Return the results
}

// QueryMaps performs a query on the database and returns the results as a slice of maps.
func QueryMaps(conn *pgx.Conn, query string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := conn.Query(context.Background(), query, args...) // Execute the query
	if err != nil {
		return nil, err // Return error if query fails
	}
	defer rows.Close() // Ensure rows are closed after processing

	var results []map[string]interface{} // Slice to hold the results

	for rows.Next() {
		fieldDescriptions := rows.FieldDescriptions()         // Get field descriptions
		values := make([]interface{}, len(fieldDescriptions)) // Create a slice to hold values
		for i := range values {
			values[i] = new(interface{}) // Initialize each element as a pointer
		}

		// Scan the row into values
		if err = rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("%w: %s", fmt.Errorf("error scanning row"), err)
		}

		row := make(map[string]interface{}) // Create a map to hold the row data
		for i, field := range fieldDescriptions {
			val := *(values[i].(*interface{})) // Dereference the value
			row[field.Name] = val              // Store the value in the map with the field name as the key
		}
		results = append(results, row) // Append the row to the results
	}

	// Check for any errors that occurred during iteration
	if err = rows.Err(); err != nil {
		return nil, err // Return error if any occurred
	}

	return results, nil // Return the results
}
