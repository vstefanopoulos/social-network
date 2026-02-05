package ct

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ------------------------------------------------------------
// Email
// ------------------------------------------------------------

// Not nullable. Error upon validation is returned if string doesn't match email format or is empty.
type Email string

func (e Email) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(e))
}

func (e *Email) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*e = Email(s)
	return nil
}

func (e Email) Validate() error {
	if !emailRegex.MatchString(string(e)) {
		return fmt.Errorf("email address must contain a single '@', have no spaces, and include a valid domain")
	}

	if err := controlCharsFree(e.String()); err != nil {
		return fmt.Errorf("email validation error: %w", err)
	}
	return nil
}

func (e Email) String() string {
	return string(e)
}

// ------------------------------------------------------------
// Username
// ------------------------------------------------------------

// Validation checks for match with usernameRegex `^[a-zA-Z0-9_]{3,32}$`
type Username string

func (u Username) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(u))
}

func (u *Username) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*u = Username(s)
	return nil
}

func (u Username) Validate() error {
	if !usernameRegex.MatchString(string(u)) {
		return fmt.Errorf("“username must be 3–32 characters long and contain only letters, numbers, or underscores”")
	}
	if err := controlCharsFree(u.String()); err != nil {
		return fmt.Errorf("username validation error: %w", err)
	}
	return nil
}

func (u Username) String() string {
	return string(u)
}

// ------------------------------------------------------------
// Identifier (username or email)
// ------------------------------------------------------------

// Represents user name or email. Identifier is a non nullable field.
// If the value doesn match username or email regexes then it is considered invalid.
type Identifier string

func (i Identifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(i))
}

func (i *Identifier) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = Identifier(s)
	return nil
}

func (i Identifier) IsValid() bool {
	s := string(i)
	return (usernameRegex.MatchString(s) || emailRegex.MatchString(s))
}

func (i Identifier) Validate() error {
	if !i.IsValid() {
		return errors.Join(ErrValidation, errors.New("identifier must be a valid username or email"))
	}
	if err := controlCharsFree(i.String()); err != nil {
		return fmt.Errorf("identifier validation error: %w", err)
	}
	return nil
}

func (i Identifier) String() string {
	return string(i)
}
