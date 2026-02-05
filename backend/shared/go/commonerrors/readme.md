# commonerrors

`commonerrors` provides a **structured, opinionated error model** for Go backend services, with a strong focus on:

* **Stable error classification**
* **Explicit error wrapping**
* **Stack trace for current error origin**
* **Clear separation between internal diagnostics and public messages**
* **Idiomatic integration with `errors.Is` / `errors.As`**
* **First-class gRPC status mapping**

This package is designed for **internal service use**, especially in gRPC-based systems, where consistent error semantics and debuggability matter more than minimal error strings.

---

## Design Goals

* Separate **classification** (what kind of error) from **cause** (what actually happened)
* Preserve the **full causal chain** of errors
* Capture **stack traces only at the point of failure**
* Support Go’s standard error inspection (`errors.Is`, `errors.As`)
* Provide a single, consistent mapping to **gRPC status codes**
* Avoid string-based error handling and ad-hoc conventions

---

## Error Classification

Error *classes* are represented as **sentinel `error` values**, similar to `os` and gRPC errors:

```go
ErrNotFound
ErrInvalidArgument
ErrInternal
ErrUnavailable
// ...
```

These sentinels are used as **error classes**, not typically returned directly.

Each error class maps 1:1 to a gRPC `codes.Code` via internal maps.

---

## The `Error` Type

```go
type Error struct {
	class      error  // Classification (never nil)
	input     string // Input / context at the wrapping site
	stack     string // Stack trace (captured only at origin)
	err       error  // Wrapped cause
	publicMsg string // Message safe for clients
}
```

### Invariants

* `class` is **guaranteed to be non-nil**
* Stack traces are captured **only at the root cause**
* Wrapping preserves the **original causal chain**
* Errors are fully compatible with `errors.Is` and `errors.As`

---

## Error String Representation

Calling `Error()` (or logging the error) produces a **recursive, flattened string** containing:

* The error class
* The input/context at each wrapping level
* The stack trace (only the origin func per linked error)
* The full wrapped error chain

```go
log.Println(err)
```

This prints the **entire causal chain**. This behavior is **intentional** and meant for **internal logging and debugging**.

## Creating Errors

### Creating a new error (origin point)

```go
err := New(ErrInternal, sqlErr, "query users")
```

* Captures a stack trace (depth 2)
* Sets the error class
* Records input/context

---

### Wrapping an existing error

```go
err = Wrap(ErrNotFound, err, "GetUser")
```

* Captures a stack trace (depth 2)
* Updates or inherits the error class
* Adds contextual input

---

### Wrapping without changing the class

```go
err = Wrap(nil, err, "repository layer")
```

* Keeps the existing classification
* Adds context only

---

### `WithCode(c error) *Error`

Updates the error’s classification code. If the provided code is invalid, the error class defaults to `ErrUnknown`.

**Parameters:**

* `c` — The new error classification.

**Returns:**
The same `*Error` instance with the updated `class`.


---

### Public-facing messages

```go
err := Wrap(ErrUnauthenticated, tokenErr, "validate token").
	WithPublic("authentication required")
```

* `publicMsg` is **never used internally**
* It is only consumed by transport adapters (e.g. gRPC)

---

### `Public() string`

Returns the public-facing message set on the error.

**Returns:**
The `publicMsg` string.

## Input & Context Formatting

Error input and contextual values are formatted using a reflection-based formatter designed for **clarity, safety, and determinism**.

### How inputs are captured

All error constructors (`New`, `Wrap`, etc.) accept arbitrary values as contextual input. These values are rendered using `FormatValue` and stored as structured text.

Each input is rendered on its own line.

---

### Named inputs

To explicitly label an input, use `Named`:

```go
err := New(
	ErrInvalid,
	Named("userID", id),
	Named("payload", req),
)
```

This produces output similar to:

```
userID = 42
payload = Request {
  Name: "Alice"
  Roles: [
    "admin"
    "editor"
  ]
}
```

Unnamed inputs are rendered as-is, one per line.

---

### Formatting rules

The formatter follows these rules:

* **Nil values**

  * Render as `nil`

* **fmt.Stringer**

  * If a value (or its pointer) implements `fmt.Stringer`, `String()` is used

* **Structs**

  * Rendered as an indented block with field names
  * Unexported fields are shown as `<unexported>`

* **Slices & arrays**

  * Rendered as indented lists

* **Maps**

  * Rendered as key-value pairs, one per line

* **Pointers**

  * Automatically dereferenced
  * Cycles are detected and rendered as `<cycle>`

* **Panics**

  * Any panic during formatting is recovered and rendered as `<unprintable>`

---

### Example

```go
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

FormatValue(
	parent{
		A: "s",
		B: 42,
		C: nestedStruct{
			a: "b",
			B: 12,
			C: map[int]int{1: 3},
		},
	},
)
```

Resulting input context:

```
parent {
   A: s
   B: 42
   C: nestedStruct {
      a: <unexported>
      B: 12
      C: map {
         1: 3
      }
   }
}
```

## Error Inspection

### Classification (`errors.Is`)

```go
if errors.Is(err, ErrNotFound) {
	// handle not found
}
```
The `Is()` custom method first checks the outer most error `commonerrors.class` then checks the wraped error. If no match it calls unwrap that looks inside the err field of each wrapped `Error`. The intention is each error to effectivelly contain only one clasification.


### Accessing the custom error (`errors.As`)

```go
var e *commonerrors.Error
if errors.As(err, &e) {
	log.Println(e)
}
```

### Root Cause Extraction

Retrieve the **original underlying error message**:

```go
cause := Source(err)
```

This walks the unwrap chain until the final error.

### Compare Class with error


A standalone helper function that checks if a given error matches a target `*Error` class.

```go
IsClass(err error, target *Error) bool
```

**Parameters:**

* `err` — The error to test.
* `target` — A reference `*Error` instance representing the target class.

**Returns:**
`true` if the error matches the target class; otherwise `false`.


### Stack
Returns the full stack and error class for each level. 

Input:
```go
package main

import commonerrors

func main(t *testing.T) {
	root := New(ErrNotFound, errors.New("sql: no rows"))
	e1 := Wrap(nil, root)
	e2 := Wrap(nil, e1)
	e3 := Wrap(ErrInternal, e2)
	stack := e3.Stack()
}
```
Resulting Output:
```
        -> commonerrors.Test_Stack at l. 9 class: internal error
        -> commonerrors.Test_Stack at l. 8 class: not found
        -> commonerrors.Test_Stack at l. 7 class: not found
        -> commonerrors.Test_Stack at l. 6 class: not found error: sql: no rows
```






## gRPC Integration

### Convert from gRPC status error

```go
DecodeProto(err error, input ...string) *Error
```

Converts a gRPC error into a `*Error` type for internal handling.

**Features:**

* Maps gRPC status codes (e.g., `NotFound`, `Internal`) to your internal error classification.
* Preserves the original gRPC error message.
* Optionally includes additional input context.
* Assigns a public-facing message using the gRPC error message.

**Parameters:**

* `err` — The gRPC error to convert.
* `input` — Optional input context strings.

**Returns:**
A new `*Error` instance representing the converted error.

---

### Convert to gRPC status error

```go
EncodeProto(err error) error
```

Converts a `*Error` (or common Go errors) back into a gRPC-compatible error.

**Features:**

* Returns gRPC errors as-is.
* Converts context errors (`DeadlineExceeded`, `Canceled`) to corresponding gRPC status codes.
* Converts domain-specific errors (`*Error`) to appropriate gRPC codes with the public message.
* Defaults to `codes.Unknown` if the error cannot be classified.

## Best Practices

* **Create errors at failure boundaries**
* **Wrap errors at layer boundaries**
* **Use `WithPublic` only for client-safe messages**


## Non-Goals

The following are intentionally out of scope:

* HTTP-specific error formatting
* Localization / i18n
* Automatic logging
* Error mutation after creation
* Structured logging output

These can be layered on top without changing the core API.


## Summary

`commonerrors` provides a **predictable, explicit, and debuggable** error model that:

* Scales cleanly across service boundaries
* Preserves full causal context
* Keeps public and internal concerns separate
* Plays nicely with Go’s standard error tooling

It favors **clarity and correctness** over minimalism, making it well-suited for production backend systems.

