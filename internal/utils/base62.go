package utils

import (
	"math/rand"
	"strings"
	"time"
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Encode a number to Base62
func toBase62(num uint) string {
	if num == 0 {
		return string(base62Alphabet[0])
	}
	var encoded strings.Builder
	for num > 0 {
		remainder := num % 62
		encoded.WriteByte(base62Alphabet[remainder])
		num /= 62
	}
	// Reverse the result since encoding builds it backward
	runes := []rune(encoded.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// GenerateURLID Generate a URL ID with random padding
func GenerateURLID(id uint, length int) string {
	base62 := toBase62(id)
	paddingLength := length - len(base62)
	if paddingLength > 0 {
		var randomPadding strings.Builder
		for i := 0; i < paddingLength; i++ {
			randomPadding.WriteByte(base62Alphabet[rand.Intn(62)])
		}
		// Mix padding and base62 for obfuscation
		return randomPadding.String() + base62
	}
	return base62
}
