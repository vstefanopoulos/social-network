package ct

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const RFC3339Layout = time.RFC3339

// ------------------------------------------------------------
// DateOfBirth
// ------------------------------------------------------------

// DateOfBirth is non nullable. If value is the zero time instant, January 1, year 1, 00:00:00 UTC validation returns error.
// It is marshaled and unmarshaled in "2006-01-02" format.
type DateOfBirth time.Time

const dobLayout = "2006-01-02"

func (d DateOfBirth) MarshalJSON() ([]byte, error) {
	t := time.Time(d)
	return json.Marshal(t.Format(dobLayout))
}

func (d *DateOfBirth) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(dobLayout, s)
	if err != nil {
		return err
	}

	*d = DateOfBirth(t)
	return nil
}

var (
	ErrDOBZero     = errors.New("date of birth is zero")
	ErrDOBInFuture = errors.New("date of birth is in the future")
	ErrDOBTooYoung = errors.New("age is below minimum allowed")
	ErrDOBTooOld   = errors.New("age exceeds maximum allowed")
)

func (d DateOfBirth) Validate() error {
	t := time.Time(d)
	if t.IsZero() {
		return ErrDOBZero
	}

	now := time.Now().UTC()

	// cannot be in the future
	if t.After(now) {
		return ErrDOBInFuture
	}

	// compute age
	age := now.Year() - t.Year()
	if now.YearDay() < t.YearDay() {
		age--
	}

	if age < dobMinAgeInYears {
		return ErrDOBTooYoung
	}

	if age > dobMaxAgeInYears {
		return ErrDOBTooOld
	}

	return nil
}

// Helper to get time.Time if needed
func (d DateOfBirth) Time() time.Time {
	return time.Time(d)
}

// Helper to parse time.Time value to proto *timestamppb.Timestamp
// If 'd' is the zero time instant returns nil
func (d DateOfBirth) ToProto() *timestamppb.Timestamp {
	if d.Time().IsZero() {
		return nil
	}
	return timestamppb.New(time.Time(d))
}

func ParseDateOfBirth(s string) (DateOfBirth, error) {
	if s == "" {
		return DateOfBirth{}, errors.New("date_of_birth is required")
	}

	t, err := time.Parse(dobLayout, s)
	if err != nil {
		return DateOfBirth{}, fmt.Errorf("invalid date_of_birth format: %w", err)
	}

	dob := DateOfBirth(t)
	if err := dob.Validate(); err != nil {
		return DateOfBirth{}, err
	}

	return dob, nil
}

func (d DateOfBirth) String() string {
	return fmt.Sprintf("%v", time.Time(d))
}

// ------------------------------------------------------------
// EventDateTime
// ------------------------------------------------------------

// It formats a time.Time value to genDateTimeLayout format.
// It Umarshals to time.Time type but Marshals to time.RFC3339 format.
//
// Null values are not allowed. If value is the zero time instant, January 1, year 1, 00:00:00 UTC validation returns error.
//
// Usage convert to proto type '*timestamppb.Timestamp':
//
//	return &pb.Event{
//			EventDateTime: resp.CreatedAt.ToProto(),
//	}, nil
type EventDateTime time.Time

func (edt EventDateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(edt)
	return json.Marshal(t.UTC().Format(RFC3339Layout))
}

func (edt *EventDateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	t, err := time.Parse(RFC3339Layout, s)
	if err != nil {
		return err
	}

	*edt = EventDateTime(t)
	return nil
}

var (
	ErrEventDateZero        = errors.New("event date/time is zero")
	ErrEventDateInPast      = errors.New("event date/time is in the past")
	ErrEventDateTooFarAhead = errors.New("event date/time exceeds maximum allowed range")
)

func (edt EventDateTime) Validate() error {
	t := time.Time(edt)
	if t.IsZero() {
		return ErrEventDateZero
	}

	now := time.Now().UTC()

	// normalize to same location
	t = t.In(now.Location())

	// must be now or later
	if t.Before(now) {
		return ErrEventDateInPast
	}

	// must not be more than N months ahead
	limit := now.AddDate(0, eventDateMaxMonthsAhead, 0)
	if t.After(limit) {
		return ErrEventDateTooFarAhead
	}

	return nil
}

// Helper to get time.Time if needed
func (edt EventDateTime) Time() time.Time {
	return time.Time(edt)
}

// Helper to parse time.Time value to proto *timestamppb.Timestamp
func (edt EventDateTime) ToProto() *timestamppb.Timestamp {
	if edt.Time().IsZero() {
		return nil
	}
	return timestamppb.New(time.Time(edt))
}

func (edt EventDateTime) String() string {
	return fmt.Sprintf("%v", time.Time(edt))
}

// ------------------------------------------------------------
// Generic Date Time
// ------------------------------------------------------------

// GenDateTime (Generic) allows null values.
// It Umarshals to time.Time type but Marshals to time.RFC3339 format.
// If is Z
//
// Usage convert to proto type '*timestamppb.Timestamp':
//
//	return &pb.Conversation{
//			CreatedAt: resp.CreatedAt.ToProto(),
//	}, nil
type GenDateTime time.Time

// Marshal to RFC3339
func (edt GenDateTime) MarshalJSON() ([]byte, error) {
	t := time.Time(edt)
	if t.IsZero() || t.Unix() == 0 {
		return []byte("null"), nil
	}
	return json.Marshal(t.UTC().Format(RFC3339Layout))
}

// Unmarshal from RFC3339 string
func (edt *GenDateTime) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" {
		*edt = GenDateTime(time.Time{})
		return nil
	}

	t, err := time.Parse(RFC3339Layout, s)
	if err != nil {
		return err
	}

	*edt = GenDateTime(t.UTC())
	return nil
}

func (edt GenDateTime) Validate() error {
	t := time.Time(edt)
	if t.IsZero() || t.Unix() == 0 {
		return fmt.Errorf("%w: zero date", ErrValidation)
	}
	return nil
}

// Scan implements the sql.Scanner interface
func (edt *GenDateTime) Scan(src any) error {
	if src == nil {
		*edt = GenDateTime(time.Time{})
		return nil
	}

	switch t := src.(type) {
	case time.Time:
		*edt = GenDateTime(t) // store exactly as DB returns it
		return nil
	case []byte:
		parsed, err := time.Parse(time.RFC3339Nano, string(t))
		if err != nil {
			return err
		}
		*edt = GenDateTime(parsed)
		return nil
	case string:
		parsed, err := time.Parse(time.RFC3339Nano, t)
		if err != nil {
			return err
		}
		*edt = GenDateTime(parsed)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into GenDateTime", src)
	}
}

// Value implements the driver.Valuer interface
func (edt GenDateTime) Value() (driver.Value, error) {
	if err := edt.Validate(); err != nil {
		return nil, fmt.Errorf("invalid GenDateTime: %w", err) // SQL NULL for invalid timestamps
	}
	return time.Time(edt), nil // store exactly as is
}

// Helper to get time.Time if needed
func (edt GenDateTime) Time() time.Time {
	return time.Time(edt)
}

// Helper to parse time.Time value to proto *timestamppb.Timestamp
func (edt GenDateTime) ToProto() *timestamppb.Timestamp {
	if edt.Time().IsZero() {
		return nil
	}

	return timestamppb.New(time.Time(edt))
}

func (g GenDateTime) String() string {
	return fmt.Sprintf("%v", time.Time(g))
}
