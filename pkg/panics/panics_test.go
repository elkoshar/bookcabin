package panics

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	SetOptions(&Options{
		Env:      "development",
		Filepath: "example",
	})
	os.Exit(m.Run())
}
