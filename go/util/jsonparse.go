package util

import (
	"encoding/json"
	"fmt"
)

type JSONParser struct {
	data map[string]interface{}
}

func NewJSONParser() *JSONParser {
	return &JSONParser{}
}

func (jp *JSONParser) Parse(jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), &jp.data)
	if err != nil {
		return err
	}
	return nil
}

func (jp *JSONParser) Get(keys ...string) (interface{}, error) {
	var targetMap map[string]interface{} = jp.data
	for _, key := range keys {
		value, ok := targetMap[key]
		if !ok {
			return nil, fmt.Errorf("key '%s' not found", key)
		}

		if m, ok := value.(map[string]interface{}); ok {
			targetMap = m
		} else {
			return value, nil
		}
	}
	return targetMap, nil
}

