package introspection

import (
	"fmt"
	"reflect"
)

// Reflection helpers for introspecting component state via struct fields.
// These are used by the Mermaid diagram generators to extract common fields
// (Name, Status, PID, Metadata, Children, etc.) from arbitrary state structs.

func getIntField(v reflect.Value, name string) int {
	field := v.FieldByName(name)
	if field.IsValid() && field.CanInt() {
		return int(field.Int())
	}
	return 0
}

func getBoolField(v reflect.Value, name string) bool {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.Bool {
		return field.Bool()
	}
	return false
}

func getStringField(v reflect.Value, name string) string {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.String {
		return field.String()
	}
	return ""
}

func getField(v reflect.Value, name string) any {
	field := v.FieldByName(name)
	if field.IsValid() && field.CanInterface() {
		return field.Interface()
	}
	return nil
}

func getMapField(v reflect.Value, name string) map[string]string {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.Map {
		result := make(map[string]string)
		iter := field.MapRange()
		for iter.Next() {
			k, v := iter.Key(), iter.Value()
			if k.Kind() == reflect.String && v.Kind() == reflect.String {
				result[k.String()] = v.String()
			}
		}
		return result
	}
	return nil
}

func getSliceField(v reflect.Value, name string) []any {
	field := v.FieldByName(name)
	if field.IsValid() && field.Kind() == reflect.Slice {
		result := make([]any, field.Len())
		for i := 0; i < field.Len(); i++ {
			result[i] = field.Index(i).Interface()
		}
		return result
	}
	return nil
}

// isNilableAndNil checks if a reflect.Value is of a nilable kind and is nil.
func isNilableAndNil(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return rv.IsNil()
	}
	return false
}

// formatReceived formats a "Received" field value for display.
func formatReceived(received any) string {
	if received == nil {
		return "None"
	}
	rv := reflect.ValueOf(received)
	if isNilableAndNil(rv) {
		return "None"
	}
	return fmt.Sprintf("%v", received)
}
