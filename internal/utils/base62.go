package utils

import (
	"fmt"
	"sync"

	"github.com/bwmarrin/snowflake"
)

// Base62 character set
const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// Singleton instance of the Snowflake node
var (
	node     *snowflake.Node
	initOnce sync.Once
)

// InitializeSnowflakeNode initializes the singleton Snowflake node
func InitializeSnowflakeNode(nodeID int64) error {
	var err error
	initOnce.Do(func() {
		node, err = snowflake.NewNode(nodeID)
	})
	return err
}

// EncodeBase62 converts an integer to a Base62 string
func EncodeBase62(num int64) string {
	if num == 0 {
		return string(base62Chars[0])
	}

	var encoded string
	for num > 0 {
		remainder := num % 62
		encoded = string(base62Chars[remainder]) + encoded
		num /= 62
	}
	return encoded
}

// GenerateShortCode generates a unique short URL code using the singleton node
func GenerateShortCode() (string, error) {
	if node == nil {
		return "", fmt.Errorf("Snowflake node is not initialized")
	}

	// Generate unique Snowflake ID
	id := node.Generate().Int64()

	// Convert the ID to a Base62 encoded string
	shortCode := EncodeBase62(id)
	return shortCode, nil
}
