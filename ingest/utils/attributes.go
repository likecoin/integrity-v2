package utils

type KeyValuePair struct {
	key   string
	value interface{}
}

func CastMapForJSON(originalMap map[interface{}]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	// handle value is array
	for key, value := range originalMap {
		if nestedArray, ok := value.([]interface{}); ok {
			stringKey := key.(string)
			newMap[stringKey] = []interface{}{}
			for _, v := range nestedArray {
				if nestedMap, ok := v.(map[interface{}]interface{}); ok {
					newMap[stringKey] = append(newMap[stringKey].([]interface{}), CastMapForJSON(nestedMap))
				} else {
					newMap[stringKey] = append(newMap[stringKey].([]interface{}), v)
				}
			}
			// handle value is map
		} else if nestedMap, ok := value.(map[interface{}]interface{}); ok {
			newMap[key.(string)] = CastMapForJSON(nestedMap)
		} else {
			newMap[key.(string)] = value
		}
	}
	return newMap
}

func ParseJsonToAttributes(jsonMap interface{}) []KeyValuePair {

	var attributes []KeyValuePair

	parsedMap, ok := jsonMap.(map[string]interface{})
	if !ok {
		return attributes
	}

	contentMetadata, ok := parsedMap["contentMetadata"]
	if !ok {
		return attributes
	}

	contentMetadataMap, ok := contentMetadata.(map[string]interface{})
	if !ok {
		return attributes
	}

	for k, v := range contentMetadataMap {
		if k != "private" {
			attributes = append(attributes, KeyValuePair{key: k, value: v})
		}
	}

	return attributes
}
