package main

import (
	"fmt"
	"strings"
)

// formatStatesDisplay formats the states string to show only first 5 states when there are more than 5
func formatStatesDisplay(states string) string {
	if states == "" {
		return ""
	}

	// Split states by comma and trim whitespace
	stateList := strings.Split(states, ",")
	for i := range stateList {
		stateList[i] = strings.TrimSpace(stateList[i])
	}

	// If 5 or fewer states, return as is
	if len(stateList) <= 5 {
		return strings.Join(stateList, ", ")
	}

	// If more than 5 states, show first 5 plus count of remaining
	firstFive := stateList[:5]
	remaining := len(stateList) - 5
	return strings.Join(firstFive, ", ") + fmt.Sprintf(", +%d more", remaining)
}

func main() {
	// Test cases
	testCases := []string{
		"",
		"California",
		"California, Nevada",
		"California, Nevada, Arizona",
		"California, Nevada, Arizona, Utah, Colorado",
		"California, Nevada, Arizona, Utah, Colorado, Wyoming",
		"California, Nevada, Arizona, Utah, Colorado, Wyoming, Montana, Idaho",
		"CA,NV,AZ,UT,CO,WY,MT,ID,OR,WA,TX,NM", // 12 states
	}

	fmt.Println("Testing formatStatesDisplay function:")
	fmt.Println("====================================")

	for i, test := range testCases {
		result := formatStatesDisplay(test)
		fmt.Printf("Test %d: Input: '%s'\n", i+1, test)
		fmt.Printf("        Output: '%s'\n", result)
		fmt.Println()
	}
}
