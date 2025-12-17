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

// SetOrderBy sets the ORDER BY clause
func (qb *QueryBuilder) SetOrderBy(sorts []SortParams) *QueryBuilder {
	if len(sorts) == 0 {
		return qb
	}

	clauses := make([]string, 0, len(sorts))
	for _, s := range sorts {
		clauses = append(clauses, s.Field+" "+strings.ToUpper(s.Direction))
	}

	qb.orderBy = "ORDER BY " + strings.Join(clauses, ", ")
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
