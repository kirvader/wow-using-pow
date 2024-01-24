package utils

import (
	"math/rand"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	t.Run("simple test", func(t *testing.T) {
		rand.Seed(123)
		result := GenerateRandomString()
		if result != "NjkwMzU=" {
			t.Errorf("GenerateRandomString returns uncontrollable result. Expected: %q, got: %q", "NjkwMzU=", result)
		}
	})
}
