package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"unicode"
)

type Headers map[string]string

const crlf = "\r\n"

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Get(key string) (string, error) {
	key = strings.ToLower(key)
	if value, exists := h[key]; exists {
		return value, nil
	} else {
		return "", fmt.Errorf("Value not found")
	}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 2, true, nil
	}
	dataString := string(data[:idx])
	dataSplits := strings.Split(dataString, ":")
	if len(dataSplits) < 2 {
		return 0, false, fmt.Errorf("Improperly formatted header %s\n", dataString)
	}
	key := dataSplits[0]
	if len(key) != len(strings.TrimRight(key, " ")) {
		return 0, false, fmt.Errorf("Invalid key formatting: %s\n", key)
	}
	key = strings.TrimSpace(key)
	if !isKeyValid(key) {
		return 0, false, fmt.Errorf("Invalid character in key: %s\n", key)
	}
	key = strings.ToLower(key)
	value := strings.Join(dataSplits[1:], ":")
	value = strings.TrimSpace(value)

	if _, exists := h[key]; exists {
		h[key] += ", " + value
	} else {
		h[key] = value
	}
	return idx + 2, false, nil
}

func (h Headers) Override(key, value string) {
	key = strings.ToLower(key)
	h[key] = value
}

func (h Headers) Remove(key string) {
	delete(h, key)
}

func isKeyValid(s string) bool {
	for _, ch := range s {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && !isSpecial(ch) {
			return false
		}
	}
	return true
}

func isSpecial(r rune) bool {
	specialChars := []rune{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}
	return slices.Contains(specialChars, r)
}
