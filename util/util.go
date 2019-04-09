package util

import (
	"fmt"
	"strconv"
)

// ChooseIndex return valid index in [0, maxLen)
func ChooseIndex(maxLen int) int {
	fmt.Print("Please choose one(index): ")
	for {
		var index string
		_, err := fmt.Scanln(&index)
		if err == nil {
			i, err := strconv.Atoi(index)
			if err == nil && i >= 0 && i < maxLen {
				return i
			}
		}
		fmt.Print("Invalid index! Please try again: ")
	}
}
