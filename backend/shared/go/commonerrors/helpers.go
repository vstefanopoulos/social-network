package commonerrors

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Returns c error if c is not nil and is a defined error
// in commonerrors else returns ErrUnknown
func parseCode(c error) error {
	if c == nil {
		c = ErrUnknown
	}
	_, ok := classToGRPC[c]
	if !ok {
		c = ErrUnknown
	}
	return c
}

// namedValue represents a value explicitly labeled with a name.
// It is used to associate structured input with a meaningful identifier
// when building error context.
type namedValue struct {
	name  string
	value any
}

// Named creates a namedValue wrapper.
//
// When passed to getInput (and ultimately error constructors),
// the name is rendered alongside the formatted value as:
//
//	<name> = <formatted value>
//
// This allows callers to explicitly label important inputs
// rather than relying on positional formatting.
func Named(name string, value any) namedValue {
	return namedValue{name: name, value: value}
}

// getInput formats a variadic list of inputs into a single string.
//
// Behavior:
//   - Each argument is rendered on its own line.
//   - If the argument is a namedValue, it is rendered as:
//     "<name> = <formatted value>"
//   - Otherwise, the argument is rendered using FormatValue directly.
//   - The final trailing newline is trimmed.
//
// This function is typically used to capture contextual input
// when creating or wrapping errors.
func getInput(args ...any) string {
	var b strings.Builder

	for _, arg := range args {
		switch v := arg.(type) {
		case namedValue:
			b.WriteString(fmt.Sprintf("%s = %s\n", v.name, FormatValue(v.value)))
		default:
			b.WriteString(FormatValue(arg))
			b.WriteString("\n")
		}
	}

	return strings.TrimRight(b.String(), "\n")
}

func getStack(depth int, skip int) string {
	var builder strings.Builder
	builder.Grow(150)
	pc := make([]uintptr, depth)
	n := runtime.Callers(skip, pc)
	if n == 0 {
		return "(no caller data)"
	}
	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)
	var count int
	for {
		count++
		frame, more := frames.Next()
		name := frame.Function
		start := strings.LastIndex(name, "/")
		builder.WriteString("level ")
		builder.WriteString(strconv.Itoa(count))
		builder.WriteString(" -> ")
		builder.WriteString(name[start+1:])
		builder.WriteString(" at l. ")
		builder.WriteString(strconv.Itoa(frame.Line))
		if !more {
			break
		}
		builder.WriteString("\n          ")
	}

	return builder.String()
}

// Helper mapper from error to grpc code.
func GetCode(err error) codes.Code {
	if err == nil {
		return codes.OK
	}

	// Propagate gRPC status errors
	if st, ok := status.FromError(err); ok {
		return st.Code()
	}

	// Handle context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return codes.DeadlineExceeded
	}
	if errors.Is(err, context.Canceled) {
		return codes.Canceled
	}

	// Handle domain error
	var e *Error
	if errors.As(err, &e) {
		if code, ok := classToGRPC[e.class]; ok {
			return code
		}
	}

	// 4. Fallback
	return codes.Unknown
}
