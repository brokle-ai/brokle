package ulid

import (
	"crypto/rand"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/oklog/ulid/v2"
)

// ULID represents a ULID type that can be used in domain models with full database support
// @Description ULID (Universally Unique Lexicographically Sortable Identifier)
// @Example "01ARZ3NDEKTSV4RRFFQ69G5FAV"
type ULID struct {
	ulid.ULID `json:"-" swaggerignore:"true"`
}

// New generates a new ULID with the current timestamp
func New() ULID {
	return ULID{ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)}
}

// NewFromTime generates a new ULID with a specific timestamp
func NewFromTime(t time.Time) ULID {
	return ULID{ulid.MustNew(ulid.Timestamp(t), rand.Reader)}
}

// Parse parses a ULID string and returns a ULID
func Parse(s string) (ULID, error) {
	parsed, err := ulid.Parse(s)
	if err != nil {
		return ULID{}, err
	}
	return ULID{parsed}, nil
}

// MustParse parses a ULID string, panicking on error
func MustParse(s string) ULID {
	parsed, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return parsed
}

// String returns the string representation of the ULID
func (u ULID) String() string {
	return u.ULID.String()
}

// Time returns the timestamp portion of the ULID
func (u ULID) Time() time.Time {
	return ulid.Time(u.ULID.Time())
}

// IsZero returns true if the ULID is zero-valued
func (u ULID) IsZero() bool {
	return u.ULID == ulid.ULID{}
}

// Scan implements the sql.Scanner interface for database reads
func (u *ULID) Scan(value interface{}) error {
	if value == nil {
		*u = ULID{}
		return nil
	}

	switch s := value.(type) {
	case string:
		parsed, err := Parse(s)
		if err != nil {
			return err
		}
		*u = parsed
		return nil
	case []byte:
		parsed, err := Parse(string(s))
		if err != nil {
			return err
		}
		*u = parsed
		return nil
	default:
		return fmt.Errorf("cannot scan %T into ULID", value)
	}
}

// Value implements the driver.Valuer interface for database writes
func (u ULID) Value() (driver.Value, error) {
	if u.IsZero() {
		return nil, nil
	}
	return u.String(), nil
}

// MarshalJSON implements the json.Marshaler interface
func (u ULID) MarshalJSON() ([]byte, error) {
	return []byte(`"` + u.String() + `"`), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (u *ULID) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		return fmt.Errorf("invalid JSON for ULID: %s", string(data))
	}

	// Remove quotes
	str := string(data[1 : len(data)-1])
	if str == "null" || str == "" {
		*u = ULID{}
		return nil
	}

	parsed, err := Parse(str)
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}

// MarshalText implements the encoding.TextMarshaler interface
func (u ULID) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface
func (u *ULID) UnmarshalText(data []byte) error {
	parsed, err := Parse(string(data))
	if err != nil {
		return err
	}
	*u = parsed
	return nil
}
