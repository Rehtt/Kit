package web

import "context"

type Option func(g *GOweb)

func WithContext(ctx context.Context) Option {
	return func(g *GOweb) {
		g.Context = ctx
	}
}
