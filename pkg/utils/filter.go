package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// FilterParams holds filter parameters extracted from query string
type FilterParams map[string]string

// FilterValidationError represents an error when invalid query parameters are provided
type FilterValidationError struct {
	InvalidParams []string
}

func (e *FilterValidationError) Error() string {
	return fmt.Sprintf("invalid query parameters: %s", strings.Join(e.InvalidParams, ", "))
}

// ParseFilters extracts allowed filter parameters from query string
// Returns error if any query parameter (except page and limit) is not in allowedParams
// Example: ParseFilters(c, []string{"name", "email", "status"})
func ParseFilters(c *gin.Context, allowedParams []string) (FilterParams, error) {
	filters := make(FilterParams)

	// Create a map of allowed params for quick lookup
	allowedMap := make(map[string]bool)
	for _, param := range allowedParams {
		allowedMap[param] = true
	}

	// Always allow pagination params
	allowedMap["page"] = true
	allowedMap["limit"] = true

	// Check all query parameters
	invalidParams := []string{}
	queryParams := c.Request.URL.Query()

	for param := range queryParams {
		if !allowedMap[param] {
			invalidParams = append(invalidParams, param)
		}
	}

	// Return error if there are invalid parameters
	if len(invalidParams) > 0 {
		return nil, &FilterValidationError{InvalidParams: invalidParams}
	}

	// Extract allowed filter parameters
	for _, param := range allowedParams {
		value := c.Query(param)
		if value != "" {
			filters[param] = value
		}
	}

	return filters, nil
}

// Get retrieves a filter value by key
func (f FilterParams) Get(key string) (string, bool) {
	value, exists := f[key]
	return value, exists
}

// Has checks if a filter key exists
func (f FilterParams) Has(key string) bool {
	_, exists := f[key]
	return exists
}

// IsEmpty checks if there are no filters
func (f FilterParams) IsEmpty() bool {
	return len(f) == 0
}
