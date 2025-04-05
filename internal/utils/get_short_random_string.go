package utils

import (
	"math/rand"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_")

// GetShortRandomString returns a random string of the specified length.
func GetShortRandomString(countOfCharacters int) string {
	countOfAlphabet := len(alphabet)
	s := make([]rune, countOfCharacters)
	for i := range countOfCharacters {
		s[i] = alphabet[rand.Intn(countOfAlphabet)]
	}

	return string(s)
}
