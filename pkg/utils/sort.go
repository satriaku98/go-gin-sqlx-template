package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type SortParams struct {
	Field     string
	Direction string
}
type SortValidationError struct {
	InvalidFields     []string
	InvalidDirections []string
}

func ParseSorts(
	c *gin.Context,
	allowedFields map[string]string,
	defaultSorts []SortParams,
) ([]SortParams, error) {

	sortQuery := c.Query("sort")
	if sortQuery == "" {
		return defaultSorts, nil
	}

	var (
		sorts             []SortParams
		invalidFields     []string
		invalidDirections []string
		used              = map[string]bool{}
	)

	parts := strings.Split(sortQuery, ",")

	for _, part := range parts {
		item := strings.Split(part, ":")
		fieldKey := item[0]

		dbField, ok := allowedFields[fieldKey]
		if !ok {
			invalidFields = append(invalidFields, fieldKey)
			continue
		}

		if used[fieldKey] {
			continue
		}
		used[fieldKey] = true

		direction := "asc"
		if len(item) == 2 {
			dir := strings.ToLower(item[1])
			if dir != "asc" && dir != "desc" {
				invalidDirections = append(invalidDirections, dir)
				continue
			}
			direction = dir
		}

		sorts = append(sorts, SortParams{
			Field:     dbField,
			Direction: direction,
		})
	}

	if len(invalidFields) > 0 || len(invalidDirections) > 0 {
		return nil, SortValidationError{
			InvalidFields:     invalidFields,
			InvalidDirections: invalidDirections,
		}
	}

	if len(sorts) == 0 {
		return defaultSorts, nil
	}

	return sorts, nil
}

func (e SortValidationError) Error() string {
	var parts []string

	if len(e.InvalidFields) > 0 {
		parts = append(parts,
			"invalid sort field: "+strings.Join(e.InvalidFields, ","),
		)
	}

	if len(e.InvalidDirections) > 0 {
		parts = append(parts,
			"invalid sort direction: "+strings.Join(e.InvalidDirections, ","),
		)
	}

	return strings.Join(parts, "\n")
}
