package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompute(t *testing.T) {
	result := Compute("hello world", "hello brave world")
	assert.NotEmpty(t, result)

	hasAdded := false
	hasUnchanged := false
	for _, l := range result {
		if l.Type == "added" {
			hasAdded = true
		}
		if l.Type == "unchanged" {
			hasUnchanged = true
		}
	}
	assert.True(t, hasAdded, "should have added lines")
	assert.True(t, hasUnchanged, "should have unchanged lines")
}

func TestCompute_Identical(t *testing.T) {
	result := Compute("same text", "same text")
	for _, l := range result {
		assert.Equal(t, "unchanged", l.Type)
	}
}

func TestCompute_Empty(t *testing.T) {
	result := Compute("", "new content")
	assert.NotEmpty(t, result)
	assert.Equal(t, "added", result[0].Type)
}
