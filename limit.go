package limiter

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Limiter struct {
	number   int
	lock     chan struct{}
	errGroup *errgroup.Group
	ctx      context.Context
}

func NewLimiter(ctx context.Context, number int) *Limiter {
	lock := make(chan struct{}, number)
	for i := 0; i < number; i++ {
		lock <- struct{}{}
	}

	group, ctx := errgroup.WithContext(ctx)

	return &Limiter{
		number:   number,
		lock:     lock,
		errGroup: group,
		ctx:      ctx,
	}
}

func (l *Limiter) push() {
	l.lock <- struct{}{}
}

func (l *Limiter) pull() {
	<-l.lock
}

func (l *Limiter) Execute(f func(ctx context.Context) error) {
	l.pull()
	l.errGroup.Go(func() error {
		defer l.push()
		return f(l.ctx)
	})
}

func (l *Limiter) Wait() error {
	for i := 0; i < l.number; i++ {
		l.pull()
	}
	return l.errGroup.Wait()
}
