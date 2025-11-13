package utility

import "github.com/bytedance/sonic"

func JSON2Map(jsonStr string) (map[string]any, error) {
	var result map[string]any
	err := sonic.Unmarshal([]byte(jsonStr), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
