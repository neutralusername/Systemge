package helpers

import "encoding/json"

// returns empty string on error
func JsonMarshal(v interface{}) string {
	json, _ := json.Marshal(v)
	return string(json)
}
