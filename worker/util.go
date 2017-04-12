package worker

import (
	"encoding/json"

	"fmt"
)

//UnmarshallCountMetric unmarshalls an array of bytes to a CountMetric
func UnmarshallCountMetric(body []byte) (*CountMetric, error) {
	var metric CountMetric
	err := json.Unmarshal(body, &metric)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse rabbit delivery body %s", err)
	}
	return &metric, nil
}

//UnmarshallCountMetricToMapInterface unmarshalls an array of bytes to a CountMetric
func UnmarshallCountMetricToMapInterface(body []byte) (map[string]interface{}, error) {
	var metric map[string]interface{}
	err := json.Unmarshal(body, &metric)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse rabbit delivery body %s", err)
	}
	return metric, nil
}
