package ct

//CT stands for Custom Types

import (
	"encoding/json"
	"fmt"
)

// Can be used for bio or descritpion.
//
// Usage:
//
//	var bioCt ct
//	var bioStr string
//	bioCt = ct.About("about me")
//	bioStr = bio.String()
type About string

func (a About) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(a))
}

func (a *About) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = About(s)
	return nil
}

func (a About) Validate() error {
	if len(a) < aboutCharsMin || len(a) > aboutCharsMax {
		return fmt.Errorf("%w: about must be %d–%d chars and contain no control characters. about length %v",
			ErrValidation,
			aboutCharsMin,
			aboutCharsMax,
			len(a),
		)
	}

	if err := controlCharsFree(a.String()); err != nil {
		return fmt.Errorf("%w, %v", ErrValidation, err)
	}

	return nil
}

func (a About) String() string {
	return string(a)
}
