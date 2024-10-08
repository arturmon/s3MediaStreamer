package postgres

import (
	"context"
	"errors"
	"fmt"
	"s3MediaStreamer/app/model"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

type UserRepositoryInterface interface {
	FindUser(ctx context.Context, value interface{}, columnType string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, email string) error
	UpdateUser(ctx context.Context, email string, fields map[string]interface{}) error
}

// FindUser retrieves a user by a specified column type and value.
func (c *Client) FindUser(ctx context.Context, value interface{}, columnType string) (model.User, error) {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "FindUser")
	defer span.End()

	var user model.User

	// Define the condition for the WHERE clause based on the column type and value.
	condition := squirrel.Eq{columnType: value}

	// Generate the SELECT query using the GenerateSelectQuery function.
	query, args := GenerateSelectQuery("users", []string{"_id", "name", "email", "password", "role", "refreshtoken", "Otp_enabled", "Otp_secret", "Otp_auth_url"}, condition)

	// Execute the query and scan the result into the user repository.
	err := c.Pool.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role,
			&user.RefreshToken, &user.OtpEnabled, &user.OtpSecret, &user.OtpAuthURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, fmt.Errorf("user not found")
		}
		return user, err
	}

	return user, nil
}

// CreateUser inserts a new user into the "users" table.
func (c *Client) CreateUser(ctx context.Context, user model.User) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "CreateUser")
	defer span.End()

	// Define the data to be inserted into the "users" table.
	userData := map[string]interface{}{
		"_id":      user.ID,
		"name":     user.Name,
		"email":    user.Email,
		"password": user.Password,
		"role":     user.Role,
		// Add other user fields as needed
	}

	// Generate the INSERT query using the GenerateInsertQuery function.
	query, args := GenerateInsertQuery("users", userData)

	// Execute the query
	_, err := c.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user by their email from the "users" table.
func (c *Client) DeleteUser(ctx context.Context, email string) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "DeleteUser")
	defer span.End()

	// Define the condition for the WHERE clause to delete the user by email.
	condition := squirrel.Eq{"email": email}

	// Generate the DELETE query using the GenerateDeleteQuery function.
	query, args := GenerateDeleteQuery("users", condition)

	// Execute the query
	result, err := c.Pool.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	// Check if any row was deleted.
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user with email '%s' not found", email)
	}

	return nil
}

// UpdateUser updates user fields in the "users" table based on the provided email.
func (c *Client) UpdateUser(ctx context.Context, email string, fields map[string]interface{}) error {
	tracer := GetTracer(ctx)
	_, span := tracer.Start(ctx, "UpdateUser")
	defer span.End()

	// Create a new instance of squirrel.UpdateBuilder
	updateBuilder := squirrel.Update("users")

	// Add SET clauses to specify the columns and their new values dynamically based on the `fields` map
	for column, value := range fields {
		updateBuilder = updateBuilder.Set(column, value)
	}

	// Add a WHERE condition to identify the user record to update based on the provided email
	updateBuilder = updateBuilder.Where(squirrel.Eq{"email": email})

	// Generate the SQL query and arguments
	sql, args, err := updateBuilder.ToSql()
	if err != nil {
		return err
	}

	// Execute the UPDATE query
	_, err = c.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}
