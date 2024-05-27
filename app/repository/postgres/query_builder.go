package postgres

import (
	"context"

	"github.com/Masterminds/squirrel"
)

// GenerateInsertQuery generates an SQL INSERT query for the specified table and data.
func GenerateInsertQuery(table string, data map[string]interface{}) (string, []interface{}) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	cols := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	for col, val := range data {
		cols = append(cols, col)
		values = append(values, val)
	}

	query, args, _ := sb.Insert(table).
		Columns(cols...).
		Values(values...).
		ToSql()

	return query, args
}

// GenerateUpdateQuery generates an SQL UPDATE query for the specified table, data, and condition.
func GenerateUpdateQuery(table string, data map[string]interface{}, condition squirrel.Sqlizer) (string, []interface{}) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	updateBuilder := sb.Update(table).SetMap(data)

	if condition != nil {
		updateBuilder = updateBuilder.Where(condition)
	}

	query, args, _ := updateBuilder.ToSql()

	return query, args
}

// GenerateSelectQuery generates an SQL SELECT query for the specified table, columns, and condition.
func GenerateSelectQuery(table string, columns []string, condition squirrel.Sqlizer) (string, []interface{}) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	selectBuilder := sb.Select(columns...).From(table)

	if condition != nil {
		selectBuilder = selectBuilder.Where(condition)
	}

	query, args, _ := selectBuilder.ToSql()

	return query, args
}

// GenerateDeleteQuery generates an SQL DELETE query for the specified table and condition.
func GenerateDeleteQuery(table string, condition squirrel.Sqlizer) (string, []interface{}) {
	sb := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	deleteBuilder := sb.Delete(table)

	if condition != nil {
		deleteBuilder = deleteBuilder.Where(condition)
	}

	query, args, _ := deleteBuilder.ToSql()

	return query, args
}

// GetTotalTrackCount returns the total count of tracks based on the query builder.
func (c *Client) GetTotalTrackCount(queryBuilder squirrel.SelectBuilder) (int, error) {
	countQuery := squirrel.Select("COUNT(*)").FromSelect(queryBuilder, "subquery")
	sql, args, err := countQuery.ToSql()
	if err != nil {
		return 0, err
	}

	var totalRows int
	err = c.Pool.QueryRow(context.TODO(), sql, args...).Scan(&totalRows)
	if err != nil {
		return 0, err
	}

	return totalRows, nil
}
