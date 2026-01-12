package helpers

func StringExists(arr []string, item string) bool {
	for i := range arr {
		if arr[i] == item {
			return true
		}
	}
	return false
}
