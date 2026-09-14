package generator

import (
	"crypto/rand"
	"errors"
	"math/big"
)

const (
	CodeLength = 10
	Alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
)

var ErrGenerationFailed = errors.New("failed to generate short code")

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate() (string, error) {
	result := make([]byte, CodeLength)
	maxInt := big.NewInt(int64(len(Alphabet)))

	for i := range result {
		n, err := rand.Int(rand.Reader, maxInt)
		if err != nil {
			return "", ErrGenerationFailed
		}

		result[i] = Alphabet[n.Int64()]
	}

	return string(result), nil
}
