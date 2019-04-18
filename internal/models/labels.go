package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"regexp"
)

type Labels map[string]interface{}

func (l Labels) Value() (driver.Value, error) {
	d, err := json.Marshal(l)
	if err != nil {
		return nil, err
	}
	return driver.Value(string(d)), nil
}

func (l *Labels) Scan(src interface{}) error {
	if src == nil {
		return fmt.Errorf("Labels.Scan: column is not nullable")
	}
	var source []byte
	switch src.(type) {
	case []byte:
		source = src.([]byte)
	case string:
		source = []byte(src.(string))
	default:
		return fmt.Errorf("Labels.Scan: Incompatible source type")
	}
	return json.Unmarshal(source, l)
}

func (l Labels) Equal(other Labels) bool {
	allEq := true
	if len(l) != len(other) {
		return false
	}
	for k, v := range l {
		if o, ok := other[k]; !ok {
			allEq = false
		} else if v != o {
			allEq = false
		}
	}
	return allEq
}

// MatchAll returns true if all labels match in other
func (l Labels) MatchAll(other Labels) bool {
	if len(l) == 0 || len(other) == 0 {
		return false
	}
	for ek, ev := range l {
		lv, ok := other[ek]
		if !ok {
			return false
		}
		var match bool
		if _, ok = lv.(string); ok {
			match, _ = regexp.MatchString(ev.(string), lv.(string))
		} else {
			match = lv == ev
		}
		if !match {
			return false
		}
	}
	return true
}

// MatchAny returns true if any label matches in other
func (l Labels) MatchAny(other Labels) bool {
	for ek, ev := range l {
		lv, ok := other[ek]
		if !ok {
			continue
		}
		var match bool
		if _, ok = lv.(string); ok {
			match, _ = regexp.MatchString(ev.(string), lv.(string))
		} else {
			match = lv == ev
		}
		if match {
			return true
		}
	}
	return false
}
