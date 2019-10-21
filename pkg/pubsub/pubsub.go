package pubsub

import (
	"context"
)

func NewPubsub() *Pubsub {
	return &Pubsub{
		themes: make(map[string][]subscriber),
	}
}

type subscriber struct {
	ctx context.Context
	ch  chan<- interface{}
}
type Pubsub struct {
	themes map[string][]subscriber
}

func (s *Pubsub) Sub(
	ctx context.Context,
	themes ...string,
) (<-chan interface{}, error) {
	ch := make(chan interface{})
	for _, t := range themes {
		err := s.makeSub(ctx, t, ch)
		if err != nil {
			return ch, err
		}
	}
	return ch, nil
}

func (s *Pubsub) makeSub(ctx context.Context, t string, ch chan<- interface{}) error {
	s.themes[t] = append(s.themes[t], subscriber{
		ctx: ctx,
		ch:  ch,
	})
	return nil
}

func (s *Pubsub) Publish(msg interface{}, themes ...string) error {
	for _, t := range themes {
		s.makePublish(msg, t)
	}
	return nil
}

func (s *Pubsub) makePublish(msg interface{}, t string) error {
	for i, sub := range s.themes[t] {
		select {
		case <-sub.ctx.Done():
			s.themes[t] = append(s.themes[t][:i], s.themes[t][:i+1]...)
		case sub.ch <- msg:
		}
	}
	return nil
}
