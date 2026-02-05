package commonerrors

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// FormatValue converts an arbitrary Go value into a readable, deterministic
// string representation suitable for error context, debugging, or logging.
//
// It is designed to be:
//   - Safe: panics during reflection are recovered and rendered as "<unprintable>"
//   - Recursive: nested structs, slices, arrays, and maps are expanded
//   - Cycle-aware: pointer cycles are detected and rendered as "<cycle>"
//   - Stringer-aware: values implementing fmt.Stringer are rendered using String()
//
// FormatValue is the public entry point. It initializes the recursion depth
// and the cycle-detection map, then delegates to formatValueIndented.
func FormatValue(v any) string {
	return formatValueIndented(v, 0, make(map[uintptr]bool))
}

// formatValueIndented recursively formats a value with indentation.
//
// Parameters:
//   - v:     the value being formatted
//   - depth: current recursion depth, used to compute indentation
//   - seen:  a map of pointer addresses used for cycle detection
//
// Behavior overview:
//
//  1. Nil handling
//     - A nil interface or nil pointer renders as "nil".
//
//  2. Interface unwrapping
//     - Interfaces are repeatedly unwrapped until a concrete value is reached.
//     - This ensures formatting is based on the underlying value, not the interface.
//
//  3. Pointer handling
//     - Nil pointers render as "nil".
//     - Non-nil pointers are tracked by address to detect cycles.
//     - Cycles render as "<cycle>" to avoid infinite recursion.
//     - The pointer is dereferenced and formatting continues on the element.
//
//  4. Stringer support
//     - If the value implements fmt.Stringer, String() is used.
//     - If the value itself does not implement Stringer but its address does,
//     the pointer receiver String() method is used.
//
//  5. Composite types
//     - Structs: rendered as a block with field names and indented values.
//     * Unexported fields are shown as "<unexported>".
//     - Maps: rendered as key-value pairs, one per line.
//     - Slices/arrays: rendered as an indexed list, one element per line.
//
//  6. Fallback
//     - All other kinds fall back to fmt.Sprintf("%v").
//
// Panic safety:
//   - Any panic encountered during reflection is recovered and rendered
//     as "<unprintable>" to avoid crashing error construction.
func formatValueIndented(v any, depth int, seen map[uintptr]bool) (out string) {
	defer func() {
		if r := recover(); r != nil {
			out = "<unprintable>"
		}
	}()

	if v == nil {
		return "nil"
	}

	if out, ok := parseProtoTime(v); ok {
		return out
	}

	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	if out, ok := implementError(val, typ); ok {
		return out
	}

	// Unwrap interfaces
	for val.Kind() == reflect.Interface {
		if val.IsNil() {
			return "nil"
		}
		val = val.Elem()
		typ = val.Type()
	}

	// Handle pointers (with cycle detection)
	if val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return "nil"
		}
		ptr := val.Pointer()
		if seen[ptr] {
			return "<cycle>"
		}
		seen[ptr] = true
		return formatValueIndented(val.Elem().Interface(), depth, seen)
	}

	if out, ok := implementStringer(val, typ); ok {
		return out
	}

	indent := strings.Repeat("   ", depth)
	nextIndent := strings.Repeat("   ", depth+1)

	switch val.Kind() {

	case reflect.Struct:
		var b strings.Builder
		name := typ.Name()
		if name == "" {
			name = "struct"
		}

		b.WriteString(name + " {\n")

		for i := 0; i < val.NumField(); i++ {
			fieldType := typ.Field(i)
			fieldVal := val.Field(i)

			b.WriteString(nextIndent + fieldType.Name + ": ")

			if fieldVal.CanInterface() {
				b.WriteString(formatValueIndented(
					fieldVal.Interface(),
					depth+1,
					seen,
				))
			} else {
				b.WriteString("<unexported>")
			}
			b.WriteString("\n")
		}

		b.WriteString(indent + "}")
		return b.String()

	case reflect.Map:
		var b strings.Builder
		b.WriteString("map {\n")

		for _, key := range val.MapKeys() {
			b.WriteString(nextIndent)
			b.WriteString(fmt.Sprintf(
				"%v: %s\n",
				key.Interface(),
				formatValueIndented(val.MapIndex(key).Interface(), depth+1, seen),
			))
		}

		b.WriteString(indent + "}")
		return b.String()

	case reflect.Slice, reflect.Array:
		var b strings.Builder
		b.WriteString("[ ")

		for i := 0; i < val.Len(); i++ {
			b.WriteString(formatValueIndented(
				val.Index(i).Interface(),
				depth+1,
				seen,
			))
			if i < val.Len()-1 {
				b.WriteString(", ")
			}
		}

		b.WriteString(indent + " ]")
		return b.String()

	default:
		return fmt.Sprintf("%v", v)
	}
}

func parseProtoTime(v any) (string, bool) {
	// Protobuf Timestamp (pointer)
	if ts, ok := v.(*timestamppb.Timestamp); ok {
		if ts == nil {
			return "nil", true
		}
		if ts.IsValid() {
			return ts.AsTime().Format(time.RFC3339), true
		}
		return "<invalid timestamp>", true
	}
	return "", false
}

func implementStringer(val reflect.Value, typ reflect.Type) (string, bool) {
	stringerType := reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

	// Value implements Stringer
	if typ.Implements(stringerType) {
		return val.Interface().(fmt.Stringer).String(), true
	}

	// Pointer implements Stringer
	if val.CanAddr() {
		ptrVal := val.Addr()
		if ptrVal.Type().Implements(stringerType) {
			return ptrVal.Interface().(fmt.Stringer).String(), true
		}
	}
	return "", false
}

func implementError(val reflect.Value, typ reflect.Type) (string, bool) {
	errorType := reflect.TypeOf((*error)(nil)).Elem()

	// Check both value and pointer receiver
	implements := typ.Implements(errorType)
	if !implements && typ.Kind() != reflect.Ptr && reflect.PointerTo(typ).Implements(errorType) {
		implements = true
	}

	if implements {
		return val.Interface().(error).Error(), true
	}
	return "", false
}
