package util

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/amsatrio/gin_notes/model/response"
)

func CamelCaseToSnakeCase(input string) string {
	regexPattern := "([a-z0-9])"
	regexOutput := "${1}"

	totalCapitalLetters := countCapitalLetters(input)
	if totalCapitalLetters > 0 {
		for i := 0; i < totalCapitalLetters; i++ {
			regexPattern = regexPattern + "([A-Z0-9])"
			regexOutput = regexOutput + "_${" + strconv.Itoa(i+2) + "}"
		}
	}

	re := regexp.MustCompile(regexPattern)
	snakeCaseString := re.ReplaceAllString(input, regexOutput)
	return strings.ToLower(snakeCaseString)
}

func countCapitalLetters(input string) int {
	count := 0

	for _, char := range input {
		if unicode.IsUpper(char) {
			count++
		}
	}

	return count
}

func ResponseToByte(response response.Response) ([]byte, error) {

	resBytes := new(bytes.Buffer)
	err := json.NewEncoder(resBytes).Encode(response)
	if err != nil {
		return nil, err
	}

	return resBytes.Bytes(), nil
}
