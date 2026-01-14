package helpers

import (
	"fmt"
	"time"
)

func StringExists(arr []string, item string) bool {
	for i := range arr {
		if arr[i] == item {
			return true
		}
	}
	return false
}

func GetTimezone(timeStr string) *time.Location {
	t, err := time.Parse(time.RFC3339, timeStr)

	if err != nil {
		layoutBatik := "2006-01-02T15:04:05-0700"
		t, err = time.Parse(layoutBatik, timeStr)
	}

	if err != nil {
		return time.FixedZone("WIB", 7*3600)
	}

	_, offset := t.Zone()
	hours := offset / 3600

	switch hours {
	case 7:
		return time.FixedZone("WIB", 7*3600) // GMT+7
	case 8:
		return time.FixedZone("WITA", 8*3600) // GMT+8
	case 9:
		return time.FixedZone("WIT", 9*3600) // GMT+9
	default:
		return t.Location()
	}
}

func FormatIDR(amount float64) string {
	intAmount := int64(amount)

	str := fmt.Sprintf("%d", intAmount)
	n := len(str)
	if n <= 3 {
		return "IDR " + str
	}

	var result []byte
	count := 0
	for i := n - 1; i >= 0; i-- {
		count++
		result = append([]byte{str[i]}, result...)
		if count == 3 && i != 0 {
			result = append([]byte{'.'}, result...)
			count = 0
		}
	}
	return "IDR " + string(result)
}
