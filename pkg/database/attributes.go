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
	argsStrs := make([]string, 0, len(args))
	for _, arg := range args {
		argsStrs = append(argsStrs, maskIfSensitive(arg.Name, arg.Value))
	}
	return []attribute.KeyValue{
		attribute.String("db.query.parameters", strings.Join(argsStrs, ", ")),
	}
}

func maskIfSensitive(name string, value any) string {
	if _, ok := sensitiveParamNames[strings.ToLower(name)]; ok {
		return "***"
	}
	return fmt.Sprintf("%v", value)
}
