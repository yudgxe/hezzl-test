package types

import (
	"database/sql"
	"encoding/json"
)

type NullString struct {
	sql.NullString
}

func (s NullString) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String)
}

func (s *NullString) UnmarshalJSON(b []byte) error {
	var v = new(string)
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}
	(*s).Valid = true
	(*s).String = *v
	return nil
}
