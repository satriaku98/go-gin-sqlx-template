package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams holds pagination parameters
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// PaginationDefaults holds default values for pagination
var PaginationDefaults = struct {
	Page  int
	Limit int
}{
	Page:  1,
	Limit: 10,
}

// ParsePagination extracts and validates pagination parameters from query string
// Returns PaginationParams with page, limit, and calculated offset
func ParsePagination(c *gin.Context) PaginationParams {
	page := parseIntParam(c, "page", PaginationDefaults.Page)
	limit := parseIntParam(c, "limit", PaginationDefaults.Limit)

	// Validate page (minimum 1)
	if page < 1 {
		page = PaginationDefaults.Page
	}

	// Validate limit (minimum 1, maximum 100)
	if limit < 1 {
		limit = PaginationDefaults.Limit
	}
	if limit > 100 {
		limit = 100
	}

	// Calculate offset
	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}

// parseIntParam parses an integer parameter from query string with a default value
func parseIntParam(c *gin.Context, key string, defaultValue int) int {
	valueStr := c.Query(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
