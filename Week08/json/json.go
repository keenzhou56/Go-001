package json

import (
	"encoding/json"
	"errors"
)

func JsonDecode(jsonStr string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b := []byte(jsonStr)
	var f interface{}
	err := json.Unmarshal(b, &f)
	if err != nil {
		err = errors.New("JsonDecode err:" + err.Error())
	} else {
		m = f.(map[string]interface{})
	}
	return m, err
}

func JsonEncode(jsonMap map[string]interface{}) ([]byte, error) {
	b, err := json.Marshal(jsonMap)
	if err != nil {
		err = errors.New("JsonEncode err:" + err.Error())
	}

	return b, err
}

/* eg
func main() {
	jsonMap := make(map[string]interface{})
	jsonMap["Name"] = "Wednesday"
	jsonMap["Age"] = 6
	jsonStr := JsonEncode(jsonMap)

	m := string(JsonDecode(jsonStr))
	fmt.Println(m["Name"], m["Age"])
}
*/
