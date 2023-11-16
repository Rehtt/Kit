package util

func DeepCopy(value any) any {
	if valueMap, ok := value.(map[string]any); ok {
		newMap := make(map[string]any)
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]any); ok {
		newSlice := make([]any, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	}

	return value
}
