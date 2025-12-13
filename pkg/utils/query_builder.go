package utils

import (
	"strings"
)

// QueryBuilder helps build SQL queries dynamically
type QueryBuilder struct {
	baseQuery    string
	whereClauses []string
	orderBy      string
	limitOffset  string
}

// NewQueryBuilder creates a new query builder with base query
func NewQueryBuilder(baseQuery string) *QueryBuilder {
	return &QueryBuilder{
		baseQuery:    baseQuery,
		whereClauses: []string{},
	}
}

// AddWhere adds a WHERE clause condition
func (qb *QueryBuilder) AddWhere(condition string) *QueryBuilder {
	if condition != "" {
		qb.whereClauses = append(qb.whereClauses, condition)
	}
	return qb
}

// AddWhereIf adds a WHERE clause condition only if the condition is met
func (qb *QueryBuilder) AddWhereIf(shouldAdd bool, condition string) *QueryBuilder {
	if shouldAdd && condition != "" {
		qb.whereClauses = append(qb.whereClauses, condition)
	}
	return qb
}

// SetOrderBy sets the ORDER BY clause
func (qb *QueryBuilder) SetOrderBy(orderBy string) *QueryBuilder {
	qb.orderBy = orderBy
	return qb
}

// SetLimitOffset sets the LIMIT and OFFSET clause
func (qb *QueryBuilder) SetLimitOffset(limit, offset string) *QueryBuilder {
	qb.limitOffset = limit + " " + offset
	return qb
}

// Build constructs the final SQL query
func (qb *QueryBuilder) Build() string {
	query := qb.baseQuery

	// Add WHERE clause
	if len(qb.whereClauses) > 0 {
		query += " WHERE " + strings.Join(qb.whereClauses, " AND ")
	}

	// Add ORDER BY
	if qb.orderBy != "" {
		query += " " + qb.orderBy
	}

	// Add LIMIT OFFSET
	if qb.limitOffset != "" {
		query += " " + qb.limitOffset
	}

	return query
}

// BuildWhereClause is a simple helper to build WHERE clause from conditions
// Returns empty string if no conditions, otherwise returns " WHERE condition1 AND condition2"
func BuildWhereClause(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(conditions, " AND ")
}

// BuildWhereClauseOr builds WHERE clause with OR operator
func BuildWhereClauseOr(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return " WHERE " + strings.Join(conditions, " OR ")
}
