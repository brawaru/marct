package terrgroup

import (
	"context"
	"golang.org/x/sync/errgroup"
)

type ThrottledErrorGroup struct {
	channel chan bool
	group   *errgroup.Group
}

func (g *ThrottledErrorGroup) Go(fc func() error) {
	g.channel <- true

	g.group.Go(func() error {
		defer func() {
			<-g.channel
		}()

		return fc()
	})
}

func (g *ThrottledErrorGroup) Wait() error {
	for i := 0; i < cap(g.channel); i++ {
		g.channel <- true
	}

	return g.group.Wait()
}

func New(concurrency int) (*ThrottledErrorGroup, context.Context) {
	return WithContext(context.Background(), concurrency)
}

func WithContext(ctx context.Context, concurrency int) (*ThrottledErrorGroup, context.Context) {
	c := make(chan bool, concurrency)

	group, cx := errgroup.WithContext(ctx)

	r := ThrottledErrorGroup{
		channel: c,
		group:   group,
	}

	return &r, cx
}
