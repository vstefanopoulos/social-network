package ct

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ------------------------------------------------------------
// Name
// ------------------------------------------------------------

// Name represents a person's first or last name. This type is non-nullable and must have a minimum length of 2 characters.
//
// Allowed characters:
// - Any Unicode letter (including letters with accents or from non-Latin alphabets, e.g., José, Łukasz, محمد)
// - Apostrophes (') and hyphens (-) within the name
// - Spaces between name parts (e.g., "Mary Jane")
//
// Restrictions:
// - Names must start and end with a Unicode letter (cannot start or end with a space, apostrophe, or hyphen)
// - Names cannot be shorter than 2 characters
// - Consecutive spaces, apostrophes, or hyphens are allowed inside the name, but the name cannot begin or end with them
// - No digits, punctuation (other than ' or -), or symbols are allowed
//
// Examples of valid names:
// "John", "O'Connor", "Jean-Luc", "Mary Jane", "José"
// Examples of invalid names:
// " John" (starts with space), "John " (ends with space), "Jo3n" (contains digit), "J@" (contains symbol)
type Name string

func (n Name) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(n))
}

func (n *Name) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*n = Name(s)
	return nil
}

func (n Name) IsValid() bool {
	if len(n) < 2 {
		return false
	}
	return nameRegex.MatchString(n.String())
}

func (n Name) Validate() error {
	if !n.IsValid() {
		return errors.New("name must be at least 2 characters long, start and end with a letter, and may only include letters, spaces, hyphens, or apostrophes; received: " + n.String())

	}

	if err := controlCharsFree(n.String()); err != nil {
		return fmt.Errorf("name type validation error: %w", err)
	}
	return nil
}

func (n Name) String() string {
	return string(n)
}
