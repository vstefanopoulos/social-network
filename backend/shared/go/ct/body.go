package ct

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ------------------------------------------------------------
// PostBody
// ------------------------------------------------------------

// Can be used for post body.
type PostBody string

func (b PostBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(b))
}

func (b *PostBody) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*b = PostBody(s)
	return nil
}

func (b PostBody) isValid() bool {
	if len(b) == 0 {
		return false
	}
	if len(b) < postBodyCharsMin || len(b) > postBodyCharsMax {
		return false
	}
	return true
}

func (b PostBody) Validate() error {
	if !b.isValid() {
		return fmt.Errorf("%w post body must be %d–%d chars and contain no control characters. post body length: %v",
			ErrValidation,
			postBodyCharsMin,
			postBodyCharsMax,
			len(b),
		)
	}

	if err := controlCharsFree(b.String()); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return nil
}

func (b PostBody) String() string {
	return string(b)
}

// ------------------------------------------------------------
// CommentBody
// ------------------------------------------------------------

// Can be used for comment body
type CommentBody string

func (c CommentBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(c))
}

func (c *CommentBody) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*c = CommentBody(s)
	return nil
}

func (c CommentBody) IsValid() bool {
	if len(c) == 0 {
		return false
	}
	if len(c) < commentBodyCharsMin || len(c) > commentBodyCharsMax {
		return false
	}

	return true
}

func (c CommentBody) Validate() error {
	if !c.IsValid() {
		return fmt.Errorf("%w comment body must be %d–%d chars and contain no control characters. comment body length: %v",
			ErrValidation,
			commentBodyCharsMin,
			commentBodyCharsMax,
			len(c),
		)
	}

	if err := controlCharsFree(c.String()); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return nil
}

func (c CommentBody) String() string {
	return string(c)
}

// ------------------------------------------------------------
// EventBody
// ------------------------------------------------------------

// Can be used for event body.
type EventBody string

func (eb EventBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(eb))
}

func (eb *EventBody) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*eb = EventBody(s)
	return nil
}

func (eb EventBody) IsValid() bool {
	if len(eb) == 0 {
		return false
	}
	if len(eb) < eventBodyCharsMin || len(eb) > eventBodyCharsMax {
		return false
	}
	return true
}

func (eb EventBody) Validate() error {
	if !eb.IsValid() {
		return fmt.Errorf("%w event body must be %d–%d chars and contain no control characters. even body length %v",
			ErrValidation,
			eventBodyCharsMin,
			eventBodyCharsMax,
			len(eb),
		)
	}
	if err := controlCharsFree(eb.String()); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return nil
}

func (eb EventBody) String() string {
	return string(eb)
}

// ------------------------------------------------------------
// EventBody
// ------------------------------------------------------------

type MsgBody string

func (m MsgBody) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(m))
}

func (m *MsgBody) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*m = MsgBody(s)
	return nil
}

func (m MsgBody) IsValid() bool {
	if len(m) == 0 {
		return false
	}
	if len(m) < msgBodyCharsMin || len(m) > msgBodyCharsMax {
		return false
	}

	return true
}

func (m MsgBody) Validate() error {
	if !m.IsValid() {
		return fmt.Errorf("%w message body must be %d–%d chars and contain no control characters. message body lenght: %v",
			ErrValidation,
			msgBodyCharsMin,
			msgBodyCharsMax,
			len(m),
		)
	}

	if err := controlCharsFree(m.String()); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}
	return nil
}

func (i *MsgBody) Scan(src any) error {
	if src == nil {
		// SQL NULL reached
		*i = ""
		return nil
	}

	switch v := src.(type) {
	case string:
		*i = MsgBody(v)
		return nil

	case []byte:
		*i = MsgBody(string(v))
		return nil
	}

	return fmt.Errorf("cannot scan type %T into MsgBody", src)
}

func (i MsgBody) Value() (driver.Value, error) {
	if !i.IsValid() {
		return nil, nil
	}
	return i.String(), nil
}

func (m MsgBody) String() string {
	return string(m)
}
