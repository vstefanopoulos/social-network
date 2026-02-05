package ct

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// ------------------------------------------------------------
// Search
// ------------------------------------------------------------

// SearchTerm represents a validated search query term
type SearchTerm string

func (s SearchTerm) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

func (s *SearchTerm) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	*s = SearchTerm(str)
	return nil
}

// Validate returns a descriptive error if the value is invalid.
func (s SearchTerm) Validate() error {
	if len(s) < 1 {
		return errors.Join(
			ErrValidation,
			errors.New("search term must be at least 1 character"),
		)
	}

	// Same regex as IsValid()
	re := regexp.MustCompile(`^[A-Za-z0-9\s\-]+$`)
	if !re.MatchString(string(s)) {
		return errors.Join(
			ErrValidation,
			errors.New("search term contains invalid characters"),
		)
	}

	if err := controlCharsFree(s.String()); err != nil {
		return fmt.Errorf("search term type validation error: %w", err)
	}

	return nil
}

func (s SearchTerm) String() string {
	return string(s)
}
