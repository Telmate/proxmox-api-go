package proxmox

import (
	"strconv"
	"strings"
)

func inArray(arr []string, str string) bool {
	for _, elem := range arr {
		if elem == str {
			return true
		}
	}

	return false
}

func Itob(i int) bool {
	if i == 1 {
		return true
	}
	return false
}

func ParseKVString(
	kvString string,
	itemsSeparator string,
	kvSeparator string,
) map[string]interface{} {
	var kvMap = map[string]interface{}{}
	var interValue interface{}
	kvStringMap := strings.Split(kvString, itemsSeparator)
	for _, item := range kvStringMap {
		if strings.Contains(item, kvSeparator) {
			itemKV := strings.Split(item, kvSeparator)
			key, value := itemKV[0], itemKV[1]
			if f, err := strconv.ParseFloat(value, 64); err == nil {
				interValue = f
			} else if i, err := strconv.ParseInt(value, 10, 64); err == nil {
				interValue = i
			} else {
				interValue = value
			}
			kvMap[key] = interValue
		}

	}
	return kvMap
}
