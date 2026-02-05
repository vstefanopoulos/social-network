package ct

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ------------------------------------------------------------
// Audience
// ------------------------------------------------------------

// Can be used for post, comment, event body.
type Audience string

func (au Audience) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(au))
}

func (au *Audience) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*au = Audience(s)
	return nil
}

func (au Audience) isValid() bool {
	if au == "" {
		return false
	}
	for _, permittedValue := range permittedAudienceValues {
		if strings.EqualFold(au.String(), permittedValue) {
			return true
		}
	}
	return false
}

func (au Audience) Validate() error {
	if !au.isValid() {
		return fmt.Errorf("%w: audience must be one of the following: %v",
			ErrValidation,
			permittedAudienceValues,
		)
	}
	if err := controlCharsFree(au.String()); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return nil
}

func (au Audience) String() string {
	return string(au)
}
