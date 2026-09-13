package generator

import (
	"strings"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	g := New()

	for range 1000 {
		code, err := g.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}

		if len(code) != CodeLength {
			t.Fatalf(
				"code length = %d, want %d",
				len(code),
				CodeLength,
			)
		}

		for _, char := range code {
			if !strings.ContainsRune(Alphabet, char) {
				t.Fatalf(
					"code contains invalid character %q",
					char,
				)
			}
		}
	}
}
