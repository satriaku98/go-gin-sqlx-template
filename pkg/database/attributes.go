package database

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/XSAM/otelsql"
	"go.opentelemetry.io/otel/attribute"
)

func GetAttrs(ctx context.Context, method otelsql.Method, query string, args []driver.NamedValue) []attribute.KeyValue {
	argsStrs := make([]string, 0, len(args))
	for _, arg := range args {
		val := fmt.Sprintf("%v", arg.Value)
		if arg.Name == "password" {
			val = "***"
		}
		argsStrs = append(argsStrs, val)
	}
	return []attribute.KeyValue{
		attribute.String("db.query.parameters", strings.Join(argsStrs, ", ")),
	}
}
