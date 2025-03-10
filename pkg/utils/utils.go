package utils

import (
	"reflect"
	"strings"
	"unicode"

	"github.com/PharmaKart/order-svc/internal/proto"
)

func ConvertMapToKeyValuePairs(m map[string]string) []*proto.KeyValuePair {
	if m == nil {
		return nil
	}

	result := make([]*proto.KeyValuePair, 0, len(m))
	for k, v := range m {
		result = append(result, &proto.KeyValuePair{
			Key:   k,
			Value: v,
		})
	}
	return result
}

// getModelColumns uses reflection to extract database column names from a GORM model
func GetModelColumns(model interface{}) map[string]bool {
	columns := make(map[string]bool)

	// Get the reflected type of the model
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	// Only process struct types
	if modelType.Kind() != reflect.Struct {
		return columns
	}

	// Iterate through all fields
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)

		// Check for embedded structs (like gorm.Model)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			// Recursively get columns from embedded struct
			embeddedColumns := GetModelColumns(reflect.New(field.Type).Elem().Interface())
			for col := range embeddedColumns {
				columns[col] = true
			}
			continue
		}

		// Get column name from GORM tag
		tag := field.Tag.Get("gorm")
		columnName := ""

		// Parse the gorm tag to find column name
		tagParts := strings.Split(tag, ";")
		for _, part := range tagParts {
			if strings.HasPrefix(part, "column:") {
				columnName = strings.TrimPrefix(part, "column:")
				break
			}
		}

		// If no column tag found, use field name converted to snake_case
		if columnName == "" {
			columnName = ToSnakeCase(field.Name)
		}

		columns[columnName] = true
	}

	// Add common GORM fields that might not be explicitly defined
	columns["id"] = true
	columns["created_at"] = true
	columns["updated_at"] = true
	columns["deleted_at"] = true

	return columns
}

// toSnakeCase converts CamelCase to snake_case
func ToSnakeCase(s string) string {
	var result string
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result += "_"
		}
		result += strings.ToLower(string(r))
	}
	return result
}
