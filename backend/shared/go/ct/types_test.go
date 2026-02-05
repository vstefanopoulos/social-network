package ct_test

import (
	"encoding/json"
	"os"
	"reflect"
	"social-network/shared/go/ct"

	"strings"
	"testing"
	"time"
)

// Utility: mustSetEnv
func mustSetEnv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}
}

// ------------------------------------------------------------
// Id
// ------------------------------------------------------------
func TestIdJSON(t *testing.T) {
	mustSetEnv(t, "ENC_KEY", "test-salt")

	id := ct.Id(123)
	b, err := json.Marshal(id)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded ct.Id
	err = json.Unmarshal(b, &decoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded != id {
		t.Fatalf("expected %d, got %d", id, decoded)
	}
}

func TestIdValidate(t *testing.T) {
	if err := ct.Id(-5).Validate(); err == nil {
		t.Fatal("expected validation error for negative Id")
	}
}

// ------------------------------------------------------------
// Id
// ------------------------------------------------------------
func TestIdValidation(t *testing.T) {
	if err := ct.Id(-1).Validate(); err == nil {
		t.Fatal("expected error for invalid id")
	}
	if err := ct.Id(5).Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ------------------------------------------------------------
// Name
// ------------------------------------------------------------
func TestNameValidation(t *testing.T) {
	if err := ct.Name("A").Validate(); err == nil {
		t.Fatal("expected name length error")
	}
	if err := ct.Name("John").Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ------------------------------------------------------------
// Username
// ------------------------------------------------------------
func TestUsernameValidation(t *testing.T) {
	if err := ct.Username("ab").Validate(); err == nil {
		t.Fatal("should fail: too short")
	}
	if err := ct.Username("valid_user123").Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ------------------------------------------------------------
// Email
// ------------------------------------------------------------
func TestEmailValidation(t *testing.T) {
	if err := ct.Email("not-an-email").Validate(); err == nil {
		t.Fatal("expected invalid email error")
	}
	if err := ct.Email("test@example.com").Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ------------------------------------------------------------
// Limit
// ------------------------------------------------------------
func TestLimitValidation(t *testing.T) {
	if err := ct.Limit(0).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := ct.Limit(500).Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if err := ct.Limit(501).Validate(); err == nil {
		t.Fatal("expected upper bound error")
	}
}

// ------------------------------------------------------------
// Offset
// ------------------------------------------------------------
func TestOffsetValidation(t *testing.T) {
	if err := ct.Offset(-1).Validate(); err == nil {
		t.Fatal("expected error for negative offset")
	}
	if err := ct.Offset(10).Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ------------------------------------------------------------
// Password
// ------------------------------------------------------------
func TestPasswordJSON(t *testing.T) {
	mustSetEnv(t, "PASSWORD_SECRET", "supersecret")

	// raw password JSON
	body := []byte(`"mySecretPass"`)

	var p ct.Password
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	// marshal must return "********"
	out, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(out) != `"********"` {
		t.Fatalf("expected masked password, got %s", out)
	}
}

func TestPasswordValidation(t *testing.T) {
	mustSetEnv(t, "PASSWORD_SECRET", "supersecret")

	var p ct.Password
	_ = json.Unmarshal([]byte(`"Password!123"`), &p)

	if err := p.Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ------------------------------------------------------------
// DateOfBirth
// ------------------------------------------------------------
func TestDOBValidation(t *testing.T) {
	now := time.Now().UTC()
	under13 := now.AddDate(-10, 0, 0)
	valid := now.AddDate(-20, 0, 0)
	future := now.AddDate(1, 0, 0)

	d := ct.DateOfBirth(valid)
	if err := d.Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	d = ct.DateOfBirth(under13)
	if err := d.Validate(); err == nil {
		t.Fatal("expected min-age error")
	}

	d = ct.DateOfBirth(future)
	if err := d.Validate(); err == nil {
		t.Fatal("expected future-date error")
	}
}

// ------------------------------------------------------------
// Identifier
// ------------------------------------------------------------
func TestIdentifierValidation(t *testing.T) {
	if err := ct.Identifier("bad@format@x").Validate(); err == nil {
		t.Fatal("expected invalid identifier")
	}
	if err := ct.Identifier("validUser_123").Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if err := ct.Identifier("email@test.com").Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

// ------------------------------------------------------------
// About
// ------------------------------------------------------------
func TestAboutValidation(t *testing.T) {
	if err := ct.About("ok!").Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if err := ct.About("\x01bad").Validate(); err == nil {
		t.Fatal("expected control char error")
	}
	if err := ct.About("ab").Validate(); err == nil {
		t.Fatal("expected min length error")
	}
}

// ------------------------------------------------------------
// Title
// ------------------------------------------------------------
func TestTitleValidation(t *testing.T) {
	if err := ct.Title("A title").Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if err := ct.Title(" ").Validate(); err == nil {
		t.Fatal("expected trimmed length error")
	}
}

// ------------------------------------------------------------
// ValidateStruct
// ------------------------------------------------------------
func TestValidateStruct(t *testing.T) {
	type TestReq struct {
		Name     ct.Name     `validate:"nullable"`
		Email    ct.Email    // required
		About    ct.About    `validate:"nullable"`
		Username ct.Username `validate:"nullable"`
		Id       ct.Id       `validate:"nullable"`
	}

	ok := TestReq{
		Name:     "John Doe",
		Email:    "valid@example.com",
		About:    "",
		Username: "user_1",
	}

	if err := ct.ValidateStruct(ok); err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	// Missing required Email
	bad := TestReq{}
	err := ct.ValidateStruct(bad)
	if err == nil {
		t.Fatal("expected missing required field error")
	}

	if !contains(err.Error(), "Email: required field missing") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// helpers
func contains(haystack, needle string) bool {
	return reflect.ValueOf(haystack).String() != "" &&
		len(haystack) >= len(needle) &&
		(len(needle) == 0 || (len(haystack) >= len(needle) && (index(haystack, needle) != -1)))
}

func index(s, sep string) int {
	return len([]rune(s[:])) - len([]rune(stringsAfter(s, sep)))
}

func stringsAfter(s, sep string) string {
	if sep == "" {
		return s
	}
	i := len([]rune(s)) - len([]rune(sep))
	if i < 0 {
		return s
	}
	return s[i:]
}

func TestValidateStruct_BoolAndOffsetExempt(t *testing.T) {
	type TestStruct struct {
		Flag   bool      `validate:""` // bool = false should NOT trigger required
		Number ct.Offset `validate:""` // Offset = 0 should NOT trigger required
		Name   ct.Name   `validate:""` // string = "" SHOULD trigger required
	}
	s := TestStruct{
		Flag:   false, // should NOT fail
		Number: 0,     // should NOT fail
		Name:   "",    // should fail
	}

	err := ct.ValidateStruct(s)
	if err == nil {
		t.Fatalf("expected validation error but got none")
	}

	// We expect ONLY Name to fail
	msg := err.Error()
	if !strings.Contains(msg, "Name: required field missing") {
		t.Fatalf("expected missing name error, got: %v", msg)
	}

	// Verify bool and Offset are NOT included in errors
	if strings.Contains(msg, "Flag") {
		t.Fatalf("bool=false should not produce error, got: %v", msg)
	}
	if strings.Contains(msg, "Number") {
		t.Fatalf("Offset=0 should not produce error, got: %v", msg)
	}
}

func TestValidateStruct_SliceOfCustomTypes(t *testing.T) {
	type TestStruct struct {
		// Slice of custom types - nullable
		NullableIDs ct.Ids `validate:"nullable"`

		// Required field to satisfy other validations
		Email ct.Email
	}

	type TestStructRequired struct {
		// Slice of custom types - not nullable
		RequiredIDs ct.Ids

		// Required field to satisfy other validations
		Email ct.Email
	}

	tests := []struct {
		name      string
		input     interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid required IDs",
			input: TestStructRequired{
				RequiredIDs: ct.Ids{1, 2},
				Email:       "test@example.com",
			},
			wantError: false,
		},
		{
			name: "nil required IDs - should fail",
			input: TestStructRequired{
				RequiredIDs: nil,
				Email:       "test@example.com",
			},
			wantError: true,
			errorMsg:  "RequiredIDs: required field missing",
		},
		{
			name: "empty required IDs - should fail",
			input: TestStructRequired{
				RequiredIDs: ct.Ids{},
				Email:       "test@example.com",
			},
			wantError: true,
			errorMsg:  "RequiredIDs: required field missing",
		},
		{
			name: "required IDs with zero element - should fail",
			input: TestStructRequired{
				RequiredIDs: ct.Ids{1, 0},
				Email:       "test@example.com",
			},
			wantError: true,
			errorMsg:  "RequiredIDs[1]: required element missing",
		},
		{
			name: "required IDs with negative element - should fail on validation",
			input: TestStructRequired{
				RequiredIDs: ct.Ids{1, -1},
				Email:       "test@example.com",
			},
			wantError: true,
			errorMsg:  "RequiredIDs[1]:",
		},
		{
			name: "nullable IDs with nil - should pass",
			input: TestStruct{
				NullableIDs: nil,
				Email:       "test@example.com",
			},
			wantError: false,
		},
		{
			name: "nullable IDs with empty slice - should pass",
			input: TestStruct{
				NullableIDs: ct.Ids{},
				Email:       "test@example.com",
			},
			wantError: false,
		},
		{
			name: "nullable IDs with zero element - should fail (elements still validated)",
			input: TestStruct{
				NullableIDs: ct.Ids{1, 0},
				Email:       "test@example.com",
			},
			wantError: true,
			errorMsg:  "NullableIDs[1]: required element missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ct.ValidateStruct(tt.input)
			if tt.wantError {
				if err == nil {
					t.Errorf("expected error on %v but got none", tt)
					return
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestValidateBatch(t *testing.T) {
	if err := ct.ValidateBatch(ct.Id(1), ct.Id(3), ct.ImgThumbnail); err != nil {
		t.Fatalf("unexpected: %v", err)
	}

	err1 := ct.ValidateBatch(ct.Id(-1), ct.Id(0), ct.Original)
	if err1 == nil {
		t.Fatal("expected error got nil")
	}
	// t.Fatal(err1)
	err2 := ct.ValidateBatch(ct.Id(-1), ct.Id(1), ct.FileVariant("invalid"))
	if err2 == nil {
		t.Fatal("expected error got nil")
	}
	// t.Fatal(err2)
}

// func TestErrorChain(t *testing.T) {
// 	// Simulate an original error
// 	origErr := errors.New("original validation failed")

// 	// Wrap it in your custom Error
// 	err1 := ct.Wrap(ct.ErrUnknownClass, origErr, "ValidateBatch step 1")

// 	// Wrap again, simulating a second validation step
// 	err2 := ct.Wrap(nil, err1, "ValidateBatch step 2")

// 	// Wrap a third time
// 	err3 := ct.Wrap(nil, err2, "ValidateBatch step 3")

// 	// Print the full error
// 	t.Fatalf("Full error chain: %v", err3.Error())

// 	// Inspecting inner errors
// 	var e *ct.Error
// 	if errors.As(err3, &e) {
// 		t.Logf("Outer error msg: %s, kind: %v", e.Msg, e.Kind)
// 		t.Logf("Inner error msg: %s, kind: %v", e.Err.(*ct.Error).Msg, e.Err.(*ct.Error).Kind)
// 	}
// }
