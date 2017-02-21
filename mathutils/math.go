package mathutils

import (
	"fmt"
	"math"
)

// CountOfDigits returns count of digits in number
func CountOfDigits(number int64) int {
	if number == 0 {
		return 1
	}
	if number < 0 {
		number = -number
	}
	return int(math.Ceil(math.Log10(math.Abs(float64(number)) + 0.5)))
}

// IntToDigits returns array of digits from number
func IntToDigits(number int64) []int {
	digits := make([]int, 0)
	if number == 0 {
		digits = append(digits, 0)
	}
	for number != 0 {
		digit := number % 10
		number = number / 10
		digits = append(digits, int(digit))
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return digits
}

// HumanBytes returns human readable bytes count.
func HumanBytes(bytes uint64) string {
	var unit uint64 = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	exp := int(math.Log(float64(bytes)) / math.Log(float64(unit)))
	pre := "KMGTPE"[exp-1 : exp]
	return fmt.Sprintf("%.2f %sB", float64(bytes)/math.Pow(float64(unit), float64(exp)), pre)
}
