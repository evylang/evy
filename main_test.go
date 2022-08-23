package main

import (
	"testing"

	"foxygo.at/evy/assert"
)

func TestTruncate(t *testing.T) {
	assert.Equal(t, "123", truncate("12345", 3))
	assert.Equal(t, "12345", truncate("12345", 5))
	assert.Equal(t, "12345", truncate("12345", 6))
	assert.Equal(t, "🦊", truncate("🦊🦊🦊🦊", 1))
}
