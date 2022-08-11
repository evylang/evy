package main

import (
	"bytes"
	"testing"

	"foxygo.at/evy/assert"
)

func TestMain(t *testing.T) {
	got := &bytes.Buffer{}
	err := compile(got)
	assert.NoError(t, err)
	assert.Equal(t, "🌱\n", got.String())
}
