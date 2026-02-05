package ct

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// ------------------------------------------------------------
// Title (group/chat title)
// ------------------------------------------------------------

// Refers to title of content (not to be confused with honorifics: mr, mrs etc). Title is a nullable field.
type Title string

func (t Title) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(t))
}

func (t *Title) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = Title(s)
	return nil
}

func (t Title) isValid() bool {
	if t == "" {
		return false
	}

	s := strings.TrimSpace(string(t))
	if len(s) < minTitleChars || len(s) > maxTitleChars {
		return false
	}

	return true
}

func (t Title) Validate() error {
	if !t.isValid() {
		return errors.Join(ErrValidation,
			fmt.Errorf("title must be %d-%d chars and contain no control characters",
				minTitleChars,
				maxTitleChars,
			))
	}

	if err := controlCharsFree(t.String()); err != nil {
		return fmt.Errorf("title type validation error: %w", err)
	}
	return nil
}

func (t Title) String() string {
	return string(t)
}
