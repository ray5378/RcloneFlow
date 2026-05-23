package runnercli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExistsStr(t *testing.T) {
	assert.False(t, existsStr(nil, "key"))
	assert.False(t, existsStr(map[string]any{}, "key"))
	assert.False(t, existsStr(map[string]any{"key": 42}, "key"))
	assert.False(t, existsStr(map[string]any{"key": true}, "key"))
	assert.True(t, existsStr(map[string]any{"key": "value"}, "key"))
	assert.True(t, existsStr(map[string]any{"key": ""}, "key"))
}

func TestExistsBool(t *testing.T) {
	assert.False(t, existsBool(nil, "key"))
	assert.False(t, existsBool(map[string]any{}, "key"))
	assert.False(t, existsBool(map[string]any{"key": "str"}, "key"))
	assert.False(t, existsBool(map[string]any{"key": 42}, "key"))
	assert.True(t, existsBool(map[string]any{"key": true}, "key"))
	assert.True(t, existsBool(map[string]any{"key": false}, "key"))
}

func TestEff(t *testing.T) {
	assert.Nil(t, eff(nil))
	assert.Nil(t, eff(map[string]any{}))
	assert.Nil(t, eff(map[string]any{"effectiveOptions": "not_a_map"}))
	assert.Equal(t, map[string]any{"k": "v"}, eff(map[string]any{"effectiveOptions": map[string]any{"k": "v"}}))
}
