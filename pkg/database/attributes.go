package database

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
)

// register sensitive parameter names
var sensitiveParamNames = map[string]struct{}{
	"password": {},
}

func GetAttrs(ctx context.Context, method otelsql.Method, query string, args []driver.NamedValue) []attribute.KeyValue {
	queryDebug := strings.TrimSpace(query)
	for _, arg := range args {
		queryDebug = strings.ReplaceAll(queryDebug, fmt.Sprintf("$%d", arg.Ordinal), maskIfSensitive(arg.Name, arg.Value))
	}
	return []attribute.KeyValue{
		attribute.String("db.query", queryDebug),
	}
}

func maskIfSensitive(name string, value any) string {
	if _, ok := sensitiveParamNames[strings.ToLower(name)]; ok {
		return "'****'"
	}
	return formatValue(value)
}

func formatValue(value any) string {
	switch v := value.(type) {
	case string:
		return fmt.Sprintf("'%s'", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}
