package helpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	var envKey = "GO_ENV"
	tests := []struct {
		name       string
		env        string
		key        string
		val        string
		fallback   string
		wantResult bool
	}{
		{
			name:       "Test Success",
			env:        EnvDevelopment,
			key:        EnvDevelopment,
			val:        EnvDevelopment,
			fallback:   EnvLocal,
			wantResult: true,
		},
		{
			name:       "Test Key not exist",
			env:        EnvDevelopment,
			key:        EnvProduction,
			val:        EnvDevelopment,
			fallback:   EnvLocal,
			wantResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv(tt.env, tt.val)
			str := GetEnv(tt.key, tt.fallback)
			if tt.wantResult {
				os.Setenv(envKey, tt.val)
				assert.Equal(t, tt.val, str)
			} else {
				os.Setenv(envKey, "")
				assert.Equal(t, tt.fallback, str)
			}

			str = GetEnvString()
			if tt.wantResult {
				assert.Equal(t, tt.val, str)
			} else {
				assert.Equal(t, EnvLocal, str)
			}
			os.Unsetenv(tt.env)
			os.Unsetenv(envKey)
		})
	}
}
