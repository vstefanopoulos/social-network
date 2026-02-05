package ct

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// ------------------------------------------------------------
// Limit
// ------------------------------------------------------------

// Non zero type. Validation returns error if zero or above limit
type Limit int32

func (l Limit) MarshalJSON() ([]byte, error) {
	return json.Marshal(int32(l))
}

func (l *Limit) UnmarshalJSON(data []byte) error {
	var v int32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*l = Limit(v)
	return nil
}

func (l Limit) isValid() bool {
	return l >= 1 && l <= Limit(maxLimit)
}

func (l Limit) Validate() error {
	if !l.isValid() {
		return errors.Join(ErrValidation, fmt.Errorf("limit must be between 1 and %d", maxLimit))
	}
	return nil
}

func (l *Limit) Scan(src any) error {
	if src == nil {
		*l = 0
		return nil
	}

	var v int64
	switch t := src.(type) {
	case int64:
		v = t
	case int32:
		v = int64(t)
	case []byte:
		n, err := strconv.ParseInt(string(t), 10, 32)
		if err != nil {
			return err
		}
		v = n
	case string:
		n, err := strconv.ParseInt(t, 10, 32)
		if err != nil {
			return err
		}
		v = n
	default:
		return fmt.Errorf("cannot scan %T into Limit", src)
	}

	*l = Limit(v)
	if !l.isValid() {
		return fmt.Errorf("invalid Limit value %d", v)
	}
	return nil
}

func (l Limit) Value() (driver.Value, error) {
	if !l.isValid() {
		return nil, fmt.Errorf("invalid Limit value %d", l)
	}
	return int64(l), nil
}

func (l Limit) Int32() int32 {
	return int32(l)
}

// ------------------------------------------------------------
// Offset
// ------------------------------------------------------------

// Non negative type, included on 'alwaysAllowZero' map when validating inside a struct.
// Validation returns error if below zero or above limit
type Offset int32

func (o Offset) MarshalJSON() ([]byte, error) {
	return json.Marshal(int32(o))
}

func (o *Offset) UnmarshalJSON(data []byte) error {
	var v int32
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*o = Offset(v)
	return nil
}

func (o Offset) isValid() bool {
	return o >= 0
}

func (o Offset) Validate() error {
	if !o.isValid() {
		return errors.Join(ErrValidation, errors.New("offset must be >= 0"))
	}
	return nil
}

func (o *Offset) Scan(src any) error {
	if src == nil {
		*o = 0
		return nil
	}

	var v int64
	switch t := src.(type) {
	case int64:
		v = t
	case int32:
		v = int64(t)
	case []byte:
		n, err := strconv.ParseInt(string(t), 10, 32)
		if err != nil {
			return err
		}
		v = n
	case string:
		n, err := strconv.ParseInt(t, 10, 32)
		if err != nil {
			return err
		}
		v = n
	default:
		return fmt.Errorf("cannot scan %T into Offset", src)
	}

	*o = Offset(v)
	if !o.isValid() {
		return fmt.Errorf("invalid Offset value %d", v)
	}
	return nil
}

func (o Offset) Value() (driver.Value, error) {
	if !o.isValid() {
		return nil, fmt.Errorf("invalid Offset value %d", o)
	}
	return int64(o), nil
}

func (o Offset) Int32() int32 {
	return int32(o)
}
