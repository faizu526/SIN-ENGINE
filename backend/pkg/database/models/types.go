package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

// JSON type for PostgreSQL JSONB
type JSON map[string]interface{}

func (j JSON) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSON) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal JSON value: %v", value)
	}
	return json.Unmarshal(bytes, j)
}

// StringArray for PostgreSQL text[]
type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	return "{" + strings.Join(a, ",") + "}", nil
}

func (a *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringArray: %v", value)
	}
	str := string(bytes)
	str = strings.Trim(str, "{}")
	if str == "" {
		*a = []string{}
		return nil
	}
	*a = strings.Split(str, ",")
	return nil
}

// IntArray for PostgreSQL integer[]
type IntArray []int

func (a IntArray) Value() (driver.Value, error) {
	str := "{"
	for i, v := range a {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("%d", v)
	}
	str += "}"
	return str, nil
}

func (a *IntArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan IntArray: %v", value)
	}
	str := string(bytes)
	str = strings.Trim(str, "{}")
	if str == "" {
		*a = []int{}
		return nil
	}
	parts := strings.Split(str, ",")
	*a = make([]int, len(parts))
	for i, p := range parts {
		fmt.Sscanf(p, "%d", &(*a)[i])
	}
	return nil
}
