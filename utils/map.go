package utils

import "sort"

// WalkMap calls callback for all values in theMap sorting keys in alphabetic order
func WalkMap[T any](theMap map[string]T, callback func(val T, key string) error) error {
	keys := make([]string, len(theMap))
	idx := 0
	for key := range theMap {
		keys[idx] = key
		idx++
	}
	sort.Strings(keys)
	for _, key := range keys {
		err := callback(theMap[key], key)
		if err != nil {
			return err
		}
	}
	return nil
}
