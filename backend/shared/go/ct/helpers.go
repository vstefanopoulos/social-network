package ct

import (
	"errors"
	"fmt"
)

var ErrControlChars error = errors.New("no control characters allowed")

// Returns error if control chars are present on 's'.
func controlCharsFree(s string) error {
	for _, r := range s {
		switch r {
		case '\n', '\r', '\t':
			continue // allowed control chars
		default:
			if r < 32 {
				return fmt.Errorf("%w: found: %v", ErrControlChars, string(r)) // reject other control chars
			}
		}
	}
	return nil
}
