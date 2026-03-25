package helpers

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// PrintDebug prints data in a formatted way for debugging purposes.
// It outputs to stdout with clear formatting, supporting maps, structs, slices, and basic types.
// Use this during development to inspect controller data.
//
// Example usage in controller:
//
//	data := core.Map{"Title": "Edit", "Product": product}
//	helpers.PrintDebug(data)
func PrintDebug(data any) {
	fmt.Println("\n========== DEBUG START ==========")

	if data == nil {
		fmt.Println("Data is nil")
		fmt.Println("=========== DEBUG END ===========\n")
		return
	}

	// Try JSON pretty print for better readability
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err == nil {
		fmt.Println(string(jsonBytes))
	} else {
		// Fallback to reflection-based printing
		printValue(reflect.ValueOf(data), 0)
	}

	fmt.Println("=========== DEBUG END ===========\n")
}

// printValue recursively prints reflect.Value with indentation
func printValue(v reflect.Value, indent int) {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	// Handle invalid values
	if !v.IsValid() {
		fmt.Printf("%s<invalid>\n", prefix)
		return
	}

	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			fmt.Printf("%s<nil>\n", prefix)
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Map:
		fmt.Printf("%s{\n", prefix)
		for _, key := range v.MapKeys() {
			fmt.Printf("%s  %v: ", prefix, key.Interface())
			mapVal := v.MapIndex(key)
			if mapVal.Kind() == reflect.Map || mapVal.Kind() == reflect.Struct || mapVal.Kind() == reflect.Slice {
				fmt.Println()
				printValue(mapVal, indent+2)
			} else {
				fmt.Printf("%v\n", mapVal.Interface())
			}
		}
		fmt.Printf("%s}\n", prefix)

	case reflect.Struct:
		t := v.Type()
		fmt.Printf("%s{\n", prefix)
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			fmt.Printf("%s  %s: ", prefix, field.Name)
			fieldVal := v.Field(i)
			if fieldVal.Kind() == reflect.Map || fieldVal.Kind() == reflect.Struct || fieldVal.Kind() == reflect.Slice {
				fmt.Println()
				printValue(fieldVal, indent+2)
			} else {
				fmt.Printf("%v\n", fieldVal.Interface())
			}
		}
		fmt.Printf("%s}\n", prefix)

	case reflect.Slice, reflect.Array:
		fmt.Printf("%s[\n", prefix)
		for i := 0; i < v.Len(); i++ {
			fmt.Printf("%s  [%d]: ", prefix, i)
			elem := v.Index(i)
			if elem.Kind() == reflect.Map || elem.Kind() == reflect.Struct || elem.Kind() == reflect.Slice {
				fmt.Println()
				printValue(elem, indent+2)
			} else {
				fmt.Printf("%v\n", elem.Interface())
			}
		}
		fmt.Printf("%s]\n", prefix)

	default:
		fmt.Printf("%s%v\n", prefix, v.Interface())
	}
}