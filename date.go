package main

import (
	"fmt"
	"time"
)

func resolveDate(input string) (string, error) {
	switch input {
	case "today":
		return time.Now().Format(time.DateOnly), nil
	case "yesterday":
		return time.Now().AddDate(0, 0, -1).Format(time.DateOnly), nil
	default:
		if _, err := time.Parse(time.DateOnly, input); err != nil {
			return "", fmt.Errorf("invalid date '%s'. Use YYYY-MM-DD, 'today', or 'yesterday'", input)
		}
		return input, nil
	}
}
