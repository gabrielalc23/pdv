package requestmeta

import "context"

type contextKey string

const metaKey contextKey = "request_meta"

func withMetadata(ctx context.Context, meta RequestMetadata) context.Context {
	return context.WithValue(ctx, metaKey, meta)
}

func FromContext(ctx context.Context) (RequestMetadata, bool) {
	meta, ok := ctx.Value(metaKey).(RequestMetadata)
	return meta, ok
}

func MustFromContext(ctx context.Context) RequestMetadata {
	meta, ok := FromContext(ctx)
	if !ok {
		return RequestMetadata{}
	}
	return meta
}
