package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type AuthRepositoryInterface interface {
	GetStoredRefreshToken(ctx context.Context, userEmail string) (string, error)
	SetStoredRefreshToken(ctx context.Context, userEmail, refreshToken string) error
}

// GetStoredRefreshToken retrieves the stored refresh token for a user by their email.
func (c *Client) GetStoredRefreshToken(ctx context.Context, userEmail string) (string, error) {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "GetStoredRefreshToken")
	defer span.End()

	var refreshToken string

	// Define the condition for the WHERE clause to select the user by email.
	condition := squirrel.Eq{"email": userEmail}

	// Generate the SELECT query using the GenerateSelectQuery function.
	query, args := GenerateSelectQuery("users", []string{"refreshtoken"}, condition)

	// Execute the query and scan the result into the refreshToken variable.
	err := c.Pool.QueryRow(ctx, query, args...).
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
func (c *Client) SetStoredRefreshToken(ctx context.Context, userEmail, refreshToken string) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "SetStoredRefreshToken")
	defer span.End()

	// Define the condition for the WHERE clause to update the user by email.
	condition := squirrel.Eq{"email": userEmail}

	// Define the data to be updated, including the new refresh token.
	updateData := map[string]interface{}{
		"refreshtoken": refreshToken,
	}

	// Generate the UPDATE query using the GenerateUpdateQuery function.
	query, args := GenerateUpdateQuery("users", updateData, condition)

	// Execute the query
	_, err := c.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
