package ct

import (
	"errors"
	"regexp"
)

type Validator interface {
	Validate() error
}

type Configs struct {
	PassSecret string
	Salt       string
}

func InitCustomTypes(PassSecret, Salt string) {
	Cfgs = Configs{
		PassSecret: PassSecret,
		Salt:       Salt,
	}
}

var Cfgs Configs

const (
	aboutCharsMin           = 3
	aboutCharsMax           = 5000
	postBodyCharsMin        = 3
	postBodyCharsMax        = 5000
	commentBodyCharsMin     = 3
	commentBodyCharsMax     = 3000
	eventBodyCharsMin       = 3
	eventBodyCharsMax       = 2000
	msgBodyCharsMin         = 1
	msgBodyCharsMax         = 1000
	dobMinAgeInYears        = 13
	dobMaxAgeInYears        = 120
	eventDateMaxMonthsAhead = 6
	maxLimit                = 500
	minTitleChars           = 1
	maxTitleChars           = 50
)

var permittedAudienceValues = []string{"everyone", "group", "followers", "selected"}

// emailRegex validates a basic email address format.
// - Requires exactly one '@' character
// - Disallows spaces anywhere in the address
// - Requires at least one character before '@'
// - Requires a domain with at least one '.' after '@'
// - Does NOT validate full RFC email rules (e.g., no quoted strings, IP literals, etc.)
var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// usernameRegex validates a simple username.
// - Allows only ASCII letters (a–z, A–Z), digits (0–9), and underscore (_)
// - Length must be between 3 and 32 characters
// - No spaces or special characters allowed
// - Must be entirely alphanumeric/underscore (no leading/trailing restrictions needed)
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)

// nameRegex validates a personal name using Unicode letters.
// - Must start with a Unicode letter
// - Must end with a Unicode letter
// - Allows internal spaces, hyphens (-), and apostrophes (')
// - Supports international (non-ASCII) characters
// - Disallows numbers, symbols, and leading/trailing whitespace or punctuation
var nameRegex = regexp.MustCompile(`^\p{L}+([\p{L}'\- ]*\p{L})?$`)

// Excluded types from nul check when validating within a struct.
var alwaysAllowZero = map[string]struct{}{
	"Offset": {},
}

var ErrValidation error = errors.New("type validation error")
