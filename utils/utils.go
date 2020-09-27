package utils

func StringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

// RemoveStringFromSlice removes the first occurrence of a from slice.
func RemoveStringFromSlice(a string, slice []string) []string {
	for i, str := range slice {
		if str == a {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// RemoveStringsFromSlice removes all occurrences of strings from slice.
func RemoveStringsFromSlice(slice []string, strings ...string) []string {
	newSlice := []string{}

	for _, item := range slice {
		if !StringInSlice(item, strings) {
			newSlice = append(newSlice, item)
		}
	}

	return newSlice
}

func RemoveDuplicatesFromStringArray(arr []string) []string {
	result := make([]string, 0, len(arr))
	seen := make(map[string]bool)

	for _, item := range arr {
		if !seen[item] {
			result = append(result, item)
			seen[item] = true
		}
	}

	return result
}
