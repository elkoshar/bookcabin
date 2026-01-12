package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	testGet(t)
	testInit(t)
	testError(t)
}

func testInit(t *testing.T) {
	// Test with no error
	err := Init(
		WithConfigFile("sample-config"),
		WithConfigFolder("../configs"),
		// WithConfigType("env"),
	)
	assert.NotNil(t, Get())
	assert.NoError(t, err)
}

func testGet(t *testing.T) {
	assert.NotNil(t, Get())
}

func testError(t *testing.T) {
	// Test with error
	err := Init(
		WithConfigFile("notfound"),
		WithConfigFolder("notfound"),
		// WithConfigType("env"),
	)
	assert.Error(t, err)
}
