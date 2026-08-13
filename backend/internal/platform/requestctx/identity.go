package requestctx

import (
	"context"
	"errors"
)

var ErrIdentityMissing = errors.New("authenticated identity missing")

type Identity struct {
	UserID   int64
	Username string
}

type identityKey struct{}

func WithIdentity(ctx context.Context, identity Identity) context.Context {
	return context.WithValue(ctx, identityKey{}, identity)
}

func IdentityFrom(ctx context.Context) (Identity, error) {
	identity, ok := ctx.Value(identityKey{}).(Identity)
	if !ok {
		return Identity{}, ErrIdentityMissing
	}
	return identity, nil
}
