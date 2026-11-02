package common

import "context"

type Actor struct {
	UserID string
	Admin  bool
}
type actorKey struct{}

// WithActor is populated only from verified authentication claims at the HTTP
// boundary. Services can enforce preview policy without depending on Gin.
func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}
func ActorFromContext(ctx context.Context) Actor {
	actor, _ := ctx.Value(actorKey{}).(Actor)
	return actor
}
