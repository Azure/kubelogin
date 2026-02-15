package options

import (
	"fmt"
	"reflect"
	"strings"
)

// ToString returns a string representation of the options for debugging
func (o *UnifiedOptions) ToString() string {
	sensitive := map[string]bool{
		"ClientSecret":       true,
		"ClientCertPassword": true,
		"Password":           true,
	}

	var parts []string
	val := reflect.ValueOf(o).Elem()
	typ := reflect.TypeOf(o).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanSet() || strings.HasPrefix(fieldType.Name, "flags") || strings.HasPrefix(fieldType.Name, "command") {
			continue
		}

		fieldValue := field.Interface()
		fieldName := fieldType.Name

		// Mask sensitive fields
		if sensitive[fieldName] {
			if str, ok := fieldValue.(string); ok && str != "" {
				fieldValue = "***"
			}
		}

		// Only include non-zero values
		if !field.IsZero() {
			parts = append(parts, fmt.Sprintf("%s: %v", fieldName, fieldValue))
		}
	}

	return fmt.Sprintf("UnifiedOptions{%s}", strings.Join(parts, ", "))
}
