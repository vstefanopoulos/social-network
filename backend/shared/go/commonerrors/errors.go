package commonerrors

import (
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Error represents a custom error type that includes classification, cause, and context.
// It implements the error interface and supports error wrapping and classification.
type Error struct {
	class     error  // Classification: ErrNotFound, ErrInternal, etc. Enusured to never be nil
	input     string // The input given to the func returning or wraping: args, structs.
	stack     string // The stack starting from the most undeliyng error and three levels up.
	err       error  // Cause: wrapped original error.
	publicMsg string // A message that will be displayed to clients.
}

// Returns a string of the full stack of errors. For each error the string contains:
//   - Error.code: Classification: ErrNotFound, ErrInternal, etc. Enusured to never be nil
//   - Error.input: The input given to the func returning or wraping: args, structs
//   - Error.err: The wraped error down the chain for higher to lower level
//   - Error.stack: The error stack
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	var builder strings.Builder
	if e.class != nil {
		builder.WriteString("\n   class: ")
		builder.WriteString(e.class.Error())
	}

	if e.input != "" {
		builder.WriteString("\n   input: ")
		builder.WriteString(e.input)
	}

	if e.stack != "" {
		builder.WriteString("\n  origin: ")
		builder.WriteString(e.stack)
	}

	if e.err != nil {
		var ce *Error
		if errors.As(e.err, &ce) {
			builder.WriteString("\n ==== Inner Custom Error: ====\n")
			builder.WriteString(e.err.Error())
		} else {
			builder.WriteString("\n   Generic Error: ")
			builder.WriteString(e.err.Error())
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// Stringer method for loggers
func (e *Error) String() string {
	return e.Error()
}

func (e *Error) Public() string {
	return e.publicMsg
}

func (e *Error) Stack() string {
	var err *Error
	if e == nil || !errors.As(e, &err) {
		return ""
	}
	var builder strings.Builder
	builder.WriteString("\n        ")
	start := strings.Index(e.stack, "-> ")
	end := strings.Index(e.stack, "\n          ")
	builder.WriteString(e.stack[start:end])
	builder.WriteString(" class: ")
	builder.WriteString(e.class.Error())
	if errors.As(e.err, &err) {
		builder.WriteString(e.err.(*Error).Stack())
	} else {
		builder.WriteString(" error: ")
		builder.WriteString(e.err.Error())
	}
	return builder.String()
}

// Returns the original most underlying error by calling Unwrap until next err is nil.
func Source(err error) string {
	for {
		u := errors.Unwrap(err)
		if u == nil {
			return err.Error()
		}
		err = u
	}
}

// Checks 'class' then 'err'.
func (e *Error) Is(target error) bool {
	// 1. Check if the target matches the classification (e.g., ErrNotFound)
	// using strict equality (standard for sentinels).
	if e.class == target {
		return true
	}

	// 2. Check if the target matches the specific Error instance pointer.
	if e == target {
		return true
	}

	// 3. Return false.
	// This signals the `errors` package to call e.Unwrap()
	// and check the underlying e.err automatically.
	return false
}

// Method for errors.IsClass parsing. Returns `Error.code` match.
func (e *Error) IsClass(target error) bool {
	return e.class == target
}

func IsClass(err error, target error) bool {
	var ce *Error
	if errors.As(target, &ce) {
		return err.(*Error).class == target
	}
	return errors.Is(err, target)
}

// Method for error.As parsing. Returns the `MediaError.Err`.
func (e *Error) Unwrap() error {
	return e.err
}

// Creates a new Error with class and optional input.
//
// Usage:
//   - kind: the classification of the error (e.g., ErrFailed, ErrNotFound). If nil, ErrUnknownClass is used.
//   - err: the underlying error to wrap; if nil, Wrap returns nil.
//   - msg: optional context message describing where or why the error occurred.
//
// Behavior:
//   - If `err` is already an Error and `kind` is nil, it preserves the original Kind and optionally adds a new message.
//   - Otherwise, it creates an new Error with the specified Kind, Err, and message.
//   - The resulting Error supports errors.Is (matches Kind) and errors.As (type assertion) and preserves the wrapped cause.
//   - If kind is nil and the err is not media error or lacks kind then kind is set to ErrUnknownClass.
//   - Logs the stack origin with depth 2 from Error creation
//
// TODO: Maybe simply call wrap ??
func New(class error, err error, input ...any) *Error {
	if err == nil {
		return nil
	}

	e := &Error{
		class: parseCode(class),
		err:   err,
		stack: getStack(2, 3),
		input: getInput(input...),
	}
	return e
}

// Wrap creates an Error that classifies and optionally wraps an existing error.
//
// Usage:
//   - kind: the classification of the error (e.g., ErrFailed, ErrNotFound). If nil, ErrUnknownClass is used.
//   - err: the underlying error to wrap; if nil, Wrap returns nil.
//   - msg: optional context message describing where or why the error occurred.
//
// Behavior:
//   - If `err` is already an Error and `kind` is nil, it preserves the original Kind and optionally adds a new message.
//   - Otherwise, it creates an new Error with the specified Kind, Err, and message.
//   - The resulting Error supports errors.Is (matches Kind) and errors.As (type assertion) and preserves the wrapped cause.
//   - If kind is nil and the err is not media error or lacks kind then kind is set to ErrUnknownClass.
//   - Logs the stack origin with depth 2 from Error creation
//
// It is recommended to only use nil kind if the underlying error is of type Error and its kind is not nil.
func Wrap(class error, err error, input ...any) *Error {
	if err == nil {
		return nil
	}

	var ce *Error
	if errors.As(err, &ce) {
		// Wrapping an existing custom error
		e := &Error{
			class:     ce.class,
			err:       err,
			stack:     getStack(2, 3),
			publicMsg: ce.publicMsg, // retain public message by default
		}

		if class != nil {
			e.class = parseCode(class)
		}
		if e.class == nil {
			e.class = ErrUnknown
		}

		e.input = getInput(input...)

		return e
	}

	e := &Error{
		class: parseCode(class),
		err:   err,
		stack: getStack(2, 3),
	}

	e.input = getInput(input...)

	return e
}

// Add a Public Message to be displayed on APIs and other public endpoints
//
// Usage:
//
//	 return Wrap(ErrUnauthorized, err, "token expired").
//		WithPublic("Authentication required")
func (e *Error) WithPublic(msg string) *Error {
	e.publicMsg = msg
	return e
}

// Returns *Error e  with error code c. If c fails validation e's code becomes ErrUnknown.
func (e *Error) WithCode(c error) *Error {
	e.class = parseCode(c)
	return e
}

// Coverts a grpc error to commonerrors Error type.
// The status code is converted to commonerrors type and the status message is wraped inside it as a new error
// as well as Error.publicMsg
//
// Optionaly a msg string is included for additional context.
// Usefull for downstream error parsing.
func DecodeProto(err error, input ...any) *Error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return New(ErrUnknown, err, getInput(input...)).WithPublic(err.Error())
	}

	code := st.Code() // codes.NotFound, codes.Internal, etc.
	message := st.Message()

	if domainErr, ok := grpcToErrorClass[code]; ok {
		return New(domainErr, errors.New(message), getInput(input...)).WithPublic(message)
	}
	return New(ErrUnknown, err, getInput(input...)).WithPublic(err.Error())
}

// Converts a commonerrors type Error to grpc status error. Handles context errors first.
// If the error passed is neither context error or Error unknown is returned.
func EncodeProto(err error) error {
	if err == nil {
		return nil
	}

	// Propagate gRPC status errors
	if st, ok := status.FromError(err); ok {
		return st.Err()
	}

	// Handle context errors
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Errorf(codes.DeadlineExceeded, "deadline exceeded")
	}
	if errors.Is(err, context.Canceled) {
		return status.Errorf(codes.Canceled, "request canceled")
	}

	// Handle domain error
	var e *Error
	if errors.As(err, &e) {
		msg := e.publicMsg
		if msg == "" {
			msg = "missing error message"
		}

		if code, ok := classToGRPC[e.class]; ok {
			return status.Errorf(code, "service error: %v", msg)
		}
	}
	return status.Errorf(codes.Unknown, "unknown error")
}
