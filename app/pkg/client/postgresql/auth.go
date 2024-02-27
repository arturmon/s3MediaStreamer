package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

// GetStoredRefreshToken retrieves the stored refresh token for a user by their email.
func (c *PgClient) GetStoredRefreshToken(userEmail string) (string, error) {
	var refreshToken string

	// Define the condition for the WHERE clause to select the user by email.
	condition := squirrel.Eq{"email": userEmail}

	// Generate the SELECT query using the GenerateSelectQuery function.
	query, args := GenerateSelectQuery("users", []string{"refreshtoken"}, condition)

	// Execute the query and scan the result into the refreshToken variable.
	err := c.Pool.QueryRow(context.TODO(), query, args...).
		Scan(&refreshToken)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("user with email '%s' not found", userEmail)
		}
		return "", err
	}

	return refreshToken, nil
}

// SetStoredRefreshToken updates the stored refresh token for a user by their email.
func (c *PgClient) SetStoredRefreshToken(userEmail, refreshToken string) error {
	// Define the condition for the WHERE clause to update the user by email.
	condition := squirrel.Eq{"email": userEmail}

	// Define the data to be updated, including the new refresh token.
	updateData := map[string]interface{}{
		"refreshtoken": refreshToken,
	}

	// Generate the UPDATE query using the GenerateUpdateQuery function.
	query, args := GenerateUpdateQuery("users", updateData, condition)

	// Execute the query
	_, err := c.Pool.Exec(context.TODO(), query, args...)
	if err != nil {
		return err
	}

	return nil
}
