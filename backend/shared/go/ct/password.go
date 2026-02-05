package ct

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// ------------------------------------------------------------
// Password
// ------------------------------------------------------------

// Password is not nullable. The length is checked and error is returned during json unmarshall and validation methods.
type Password string

func (p Password) MarshalJSON() ([]byte, error) {
	// No encoder required – return placeholder or omit
	return json.Marshal("********")
}

func (p *Password) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) == 0 {
		return errors.Join(ErrValidation, errors.New("password is required"))
	}

	*p = Password(raw)
	return nil
}

func (p Password) Hash() (Password, error) {
	secret := Cfgs.PassSecret

	if secret == "" {
		return "", errors.Join(ErrValidation, errors.New("missing env var PASSWORD_SECRET"))
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(p))
	p = Password(base64.StdEncoding.EncodeToString(mac.Sum(nil)))
	return p, nil
}

// one symbol, one capital letter, one number min 8 chars max 64 chars
var (
	uppercase = regexp.MustCompile(`[A-Z]`)
	lowercase = regexp.MustCompile(`[a-z]`)
	digit     = regexp.MustCompile(`[0-9]`)
	symbol    = regexp.MustCompile(`[^A-Za-z0-9]`)
)

var (
	ErrPasswordTooShort     = errors.New("password is too short")
	ErrPasswordTooLong      = errors.New("password is too long")
	ErrPasswordNoUppercase  = errors.New("password must contain an uppercase letter")
	ErrPasswordNoLowercase  = errors.New("password must contain a lowercase letter")
	ErrPasswordNoDigit      = errors.New("password must contain a digit")
	ErrPasswordNoSymbol     = errors.New("password must contain a symbol")
	ErrPasswordControlChars = errors.New("password contains control characters")
)

func (p Password) Validate() error {
	s := string(p)

	if len(s) < 8 {
		return ErrPasswordTooShort
	}
	if len(s) > 64 {
		return ErrPasswordTooLong
	}
	if !uppercase.MatchString(s) {
		return ErrPasswordNoUppercase
	}
	if !lowercase.MatchString(s) {
		return ErrPasswordNoLowercase
	}
	if !digit.MatchString(s) {
		return ErrPasswordNoDigit
	}
	if !symbol.MatchString(s) {
		return ErrPasswordNoSymbol
	}
	if err := controlCharsFree(p.String()); err != nil {
		return err
	}

	return nil
}

func (p Password) String() string {
	return string(p)
}

// ------------------------------------------------------------
// HashedPassword
// ------------------------------------------------------------

// HashedPassword is not nullable. The length is checked and error is returned during json unmarshall and validation methods.
type HashedPassword string

func (hp HashedPassword) MarshalJSON() ([]byte, error) {
	// No encoder required – return placeholder or omit
	return json.Marshal("********")
}

func (hp *HashedPassword) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) == 0 {
		return errors.Join(ErrValidation, errors.New("password is required"))
	}

	*hp = HashedPassword(raw)
	return nil
}

func (hp HashedPassword) isValid() bool {
	return hp != ""
}

func (hp HashedPassword) Validate() error {
	if !hp.isValid() {
		return errors.Join(ErrValidation, errors.New("invalid hashed password"))
	}
	if err := controlCharsFree(hp.String()); err != nil {
		return fmt.Errorf("%w, %v", ErrValidation, err)
	}
	return nil
}

func (hp HashedPassword) String() string {
	return string(hp)
}
