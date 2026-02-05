package commonerrors

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "nil kind defaults to ErrUnknownClass",
			err:      &Error{class: nil, input: "test"},
			expected: "\n   input: test",
		},
		{
			name:     "kind, msg, and wrapped error",
			err:      &Error{class: ErrNotFound, input: "user not found", err: errors.New("db error")},
			expected: "\n   class: not found\n   input: user not found\n   Generic Error: db error\n",
		},
		{
			name:     "kind and msg only",
			err:      &Error{class: ErrInternal, input: "internal error"},
			expected: "\n   class: internal error\n   input: internal error",
		},
		{
			name:     "kind and wrapped error only",
			err:      &Error{class: ErrInvalidArgument, err: errors.New("invalid input")},
			expected: "\n   class: invalid argument\n   Generic Error: invalid input\n",
		},
		{
			name:     "kind only",
			err:      &Error{class: ErrPermissionDenied},
			expected: "\n   class: permission denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestError_Is(t *testing.T) {
	err := &Error{class: ErrNotFound}

	assert.True(t, errors.Is(err, ErrNotFound))
	assert.False(t, errors.Is(err, ErrInternal))
}

func TestError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &Error{err: underlying}

	assert.Equal(t, underlying, errors.Unwrap(err))
}

func TestNew(t *testing.T) {
	err := New(nil, root())
	assert.Equal(t, `
   class: unknown error
  origin: level 1 -> commonerrors.TestNew at l. 71
          level 2 -> testing.tRunner at l. 1934
   Generic Error: sql: no rows
`, err.Error())
}

func TestWrap(t *testing.T) {
	underlying := errors.New("underlying")

	t.Run("wrap nil returns nil", func(t *testing.T) {
		assert.Nil(t, Wrap(ErrInternal, nil))
	})

	t.Run("wrap with kind and msg", func(t *testing.T) {
		err := Wrap(ErrNotFound, underlying, "not found")
		require.NotNil(t, err)
		assert.Equal(t, ErrNotFound, err.class)
		assert.Equal(t, underlying, err.err)
		assert.Equal(t, "not found", err.input)
		assert.Equal(t, "", err.publicMsg)
	})

	t.Run("wrap existing Error with new kind and msg", func(t *testing.T) {
		existing := &Error{class: ErrInternal, err: underlying, publicMsg: "public"}
		wrapped := Wrap(ErrNotFound, existing, "updated")
		require.NotNil(t, wrapped)
		assert.Equal(t, ErrNotFound, wrapped.class)
		assert.Equal(t, existing, wrapped.err)
		assert.Equal(t, "updated", wrapped.input)
		assert.Equal(t, "public", wrapped.publicMsg)
	})

	t.Run("wrap existing Error preserving kind when new kind nil", func(t *testing.T) {
		existing := &Error{class: ErrInternal, err: underlying}
		wrapped := Wrap(nil, existing, "preserved")
		require.NotNil(t, wrapped)
		assert.Equal(t, ErrInternal, wrapped.class)
		assert.Equal(t, existing, wrapped.err)
		assert.Equal(t, "preserved", wrapped.input)
	})

	t.Run("wrap with nil kind defaults to ErrUnknownClass", func(t *testing.T) {
		err := Wrap(nil, underlying)
		require.NotNil(t, err)
		assert.Equal(t, ErrUnknown, err.class)
	})
}

func TestWithPublic(t *testing.T) {
	err := &Error{class: ErrInternal}
	result := err.WithPublic("public message")

	assert.Equal(t, err, result) // returns same instance
	assert.Equal(t, "public message", err.publicMsg)
}

func TestGetCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected codes.Code
	}{
		{name: "nil error", err: nil, expected: codes.OK},
		{name: "ErrOK", err: &Error{class: ErrOK}, expected: codes.OK},
		{name: "ErrNotFound", err: &Error{class: ErrNotFound}, expected: codes.NotFound},
		{name: "ErrInternal", err: &Error{class: ErrInternal}, expected: codes.Internal},
		{name: "ErrInvalidArgument", err: &Error{class: ErrInvalidArgument}, expected: codes.InvalidArgument},
		{name: "ErrAlreadyExists", err: &Error{class: ErrAlreadyExists}, expected: codes.AlreadyExists},
		{name: "ErrPermissionDenied", err: &Error{class: ErrPermissionDenied}, expected: codes.PermissionDenied},
		{name: "ErrUnauthenticated", err: &Error{class: ErrUnauthenticated}, expected: codes.Unauthenticated},
		{name: "unknown kind", err: &Error{class: errors.New("custom")}, expected: codes.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetCode(tt.err))
		})
	}
}

func TestIntegration(t *testing.T) {
	underlying := errors.New("db connection failed")
	customErr := Wrap(ErrInternal, underlying, "failed to connect")

	// Test errors.Is works
	assert.True(t, errors.Is(customErr, underlying))
	assert.True(t, errors.Is(customErr, ErrInternal))
	assert.False(t, errors.Is(customErr, ErrNotFound))

	// Test errors.As works
	var target *Error
	assert.True(t, errors.As(customErr, &target))
	assert.Equal(t, ErrInternal, target.class)

	// Test unwrapping chain
	assert.Equal(t, underlying, errors.Unwrap(customErr))
}

func TestMultiLayerWrap_ErrorString(t *testing.T) {
	root := root()
	e1 := err1(root)
	e2 := Err2(e1)
	e3 := err3(e2)

	out := e3.Error()
	// t.Fatal(out)

	assert.Contains(t, out, "level 3")
	assert.Contains(t, out, "level 2")
	assert.Contains(t, out, "level 1")
	assert.Contains(t, out, "sql: no rows")

	// stack appears exactly once (entire block)
	assert.Equal(t, 1, strings.Count(out, e1.(*Error).stack))
}

func root() error {
	return errors.New("sql: no rows")
}
func err1(err error) error {
	return New(ErrNotFound, err, "level 1")
}
func Err2(err error) error {
	return Wrap(nil, err, "level 2")
}
func err3(err error) error {
	return Wrap(ErrInternal, err, "level 3")
}

func TestAs_ReturnsOutermostError(t *testing.T) {
	root := errors.New("io failure")
	e1 := New(ErrUnavailable, root, "dial")
	e2 := Wrap(nil, e1, "retry")

	var ce *Error
	require.True(t, errors.As(e2, &ce))
	assert.Equal(t, ErrUnavailable, ce.class)
	assert.Equal(t, e2, ce)
}

func TestGetSource_MultiWrap(t *testing.T) {
	root := errors.New("disk full")
	err := Wrap(
		ErrInternal,
		Wrap(nil,
			New(ErrUnavailable, root, "storage"),
			"service",
		),
		"handler",
	)

	assert.Equal(t, "disk full", Source(err))
}

func TestEncodeProto_DoesNotMutateError(t *testing.T) {
	root := errors.New("token expired")
	err := Wrap(ErrUnauthenticated, root, "auth").
		WithPublic("authentication required")

	_ = EncodeProto(err)

	// publicMsg must remain unchanged
	var ce *Error
	require.True(t, errors.As(err, &ce))
	assert.Equal(t, "authentication required", ce.publicMsg)
}

// Ensure most outer code prevails over nested codes
func TestEncodeProto_MultipleWrapedCodes(t *testing.T) {
	root := New(ErrUnknown, errors.New("token expired"))
	err := Wrap(ErrUnauthenticated, root, "auth")

	out := EncodeProto(err)
	require.NotNil(t, out)

	st, ok := status.FromError(out)
	require.True(t, ok, "expected gRPC status error")

	assert.Equal(t, codes.Unauthenticated, st.Code())
}

func TestEncodeProto_DefaultPublicMessage(t *testing.T) {
	err := Wrap(ErrInternal, errors.New("panic"), "handler")

	st, ok := status.FromError(EncodeProto(err))
	require.True(t, ok)

	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "missing error message")
}

func Test_ErrorFormating(t *testing.T) {
	err := New(ErrInternal, errors.New("panic"), "handler")
	s := err.Error()
	assert.Equal(t, `
   class: internal error
   input: handler
  origin: level 1 -> commonerrors.Test_ErrorFormating at l. 266
          level 2 -> testing.tRunner at l. 1934
   Generic Error: panic
`, s)
}

func Test_Stack(t *testing.T) {
	root := New(ErrNotFound, errors.New("sql: no rows"))
	e1 := Wrap(nil, root)
	e2 := Wrap(nil, e1)
	e3 := Wrap(ErrInternal, e2)
	stack := e3.Stack()
	assert.Equal(t, `
        -> commonerrors.Test_Stack at l. 281 class: internal error
        -> commonerrors.Test_Stack at l. 280 class: not found
        -> commonerrors.Test_Stack at l. 279 class: not found
        -> commonerrors.Test_Stack at l. 278 class: not found error: sql: no rows`, stack)
}

func Test_Source(t *testing.T) {
	root := New(ErrNotFound, errors.New("sql: no rows"))
	e1 := Wrap(nil, root)
	e2 := Wrap(nil, e1)
	e3 := Wrap(ErrInternal, e2)
	stack := Source(e3)
	assert.Equal(t, `sql: no rows`, stack)
}

func TestGetInputFormatting(t *testing.T) {
	type Sample struct {
		A int
		B string
	}

	tests := []struct {
		name     string
		input    []any
		expected string
	}{
		{
			name: "single named primitive",
			input: []any{
				Named("argA", 42),
			},
			expected: "argA = 42",
		},
		{
			name: "multiple named primitives",
			input: []any{
				Named("argA", 1),
				Named("argB", "hello"),
			},
			expected: "argA = 1\nargB = hello",
		},
		{
			name: "struct formatting",
			input: []any{
				Sample{A: 10, B: "x"},
			},
			expected: strings.TrimSpace(`
Sample {
   A: 10
   B: x
}`),
		},
		{
			name: "named struct",
			input: []any{
				Named("payload", Sample{A: 1, B: "y"}),
			},
			expected: strings.TrimSpace(`
payload = Sample {
   A: 1
   B: y
}`),
		},
		{
			name: "map formatting",
			input: []any{
				map[string]int{"a": 1},
			},
			expected: strings.TrimSpace(`
map {
   a: 1
}`),
		},
		{
			name: "slice formatting",
			input: []any{
				[]string{"x", "y"},
			},
			expected: strings.TrimSpace(`
[ x, y ]`),
		},
		{
			name: "mixed inputs",
			input: []any{
				Named("id", 7),
				Sample{A: 2, B: "z"},
				[]int{1, 2},
			},
			expected: strings.TrimSpace(`
id = 7
Sample {
   A: 2
   B: z
}
[ 1, 2 ]`),
		},
		{
			name: "slice of structs",
			input: []any{
				Sample{A: 1, B: "one"},
				Sample{A: 2, B: "two"},
				Sample{A: 3, B: "three"},
			},
			expected: strings.TrimSpace(`
Sample {
   A: 1
   B: one
}
Sample {
   A: 2
   B: two
}
Sample {
   A: 3
   B: three
}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := getInput(tt.input...)
			if out != tt.expected {
				t.Fatalf("unexpected output:\n--- got ---\n%s\n--- want ---\n%s",
					out, tt.expected)
			}
		})
	}
}

func TestWrapCapturesFormattedInput(t *testing.T) {
	type Req struct {
		UserID string
		Count  int
	}

	baseErr := errors.New("db failure")
	s := "u123"
	e := Wrap(
		ErrInternal,
		baseErr,
		Named("userID", s),
		Req{UserID: "u123", Count: 3},
		map[string]int{"x": 9},
	)

	if e == nil {
		t.Fatal("expected non-nil error")
	}

	expected := strings.TrimSpace(`
userID = u123
Req {
   UserID: u123
   Count: 3
}
map {
   x: 9
}`)

	if e.input != expected {
		t.Fatalf("input formatting mismatch:\n--- got ---\n%s\n--- want ---\n%s",
			e.input, expected)
	}
}

func TestNestedStruct(t *testing.T) {
	type nestedStruct struct {
		a string
		B int
		C map[int]int
	}
	type parent struct {
		A string
		B int
		C nestedStruct
	}

	n := FormatValue(
		parent{
			A: "s",
			B: 42,
			C: nestedStruct{
				a: "b",
				B: 12,
				C: map[int]int{1: 3},
			},
		})
	assert.Equal(t,
		`parent {
   A: s
   B: 42
   C: nestedStruct {
      a: <unexported>
      B: 12
      C: map {
         1: 3
      }
   }
}`, n)
}

func TestProtobufFormating(t *testing.T) {

	protoStamp := timestamppb.Now()
	stamp := protoStamp.AsTime().Format(time.RFC3339)
	n := FormatValue(protoStamp)
	assert.Equal(t, n, stamp)
}

func TestErrFormating(t *testing.T) {
	err := errors.New("error")
	n := FormatValue(err)
	assert.Equal(t, err.Error(), n)
}
