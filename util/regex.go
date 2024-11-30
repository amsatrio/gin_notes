package util

import "regexp"

func ValidateLetterNumber(input string) bool {
	regex := regexp.MustCompile(`^[a-zA-Z0-9\s]+$`)
	return regex.MatchString(input)
}

func ValidateEmail(input string) bool {
	regex := regexp.MustCompile(`^[\w\-\.]+@([\w-]+\.)+[\w-]{2,}$`)
	return regex.MatchString(input)
}
