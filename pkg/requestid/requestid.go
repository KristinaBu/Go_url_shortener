package requestid

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type contextKey struct{}

func New() string {
	buffer := make([]byte, 16)
	_, _ = rand.Read(buffer)

	return hex.EncodeToString(buffer)
}

func WithID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}
