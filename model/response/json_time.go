package response

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type JSONTime struct {
	time.Time
}

func (jt JSONTime) MarshalJSON() ([]byte, error) {
	formatted := jt.Format("2006-01-02 15:04:05")
	return []byte(`"` + formatted + `"`), nil
}

func (jt *JSONTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)

	parsedTime, err := time.Parse("2006-01-02 15:04:05", str)
	if err != nil {
		return err
	}

	jt.Time = parsedTime

	return nil
}

func (jt *JSONTime) Scan(value interface{}) error {
	if value == nil {
		*jt = JSONTime{Time: time.Time{}}
		return nil
	}

	// t, ok := value.(time.Time)
	switch st := value.(type) {
	case time.Time:
		*jt = JSONTime{Time: st}
	case []byte:
		return json.Unmarshal(st, jt)
	case string:
		return json.Unmarshal([]byte(st), jt)
	default:
		return errors.New("unsupported type for JSONTime")
	}
	return nil
}

func (jt JSONTime) Value() (driver.Value, error) {
	if jt.IsZero() {
		return nil, nil
	}
	return jt.Time, nil
}
