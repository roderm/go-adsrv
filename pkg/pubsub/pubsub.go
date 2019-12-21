package pubsub

import (
	"context"
)

// NewPubsub Publisher/Subscribe object
func NewPubsub() *Pubsub {
	return &Pubsub{
		topics: make(map[string][]subscriber),
	}
}

type subscriber struct {
	ctx context.Context
	ch  chan<- interface{}
}
type Pubsub struct {
	topics map[string][]subscriber
}

// Sub subscribe to multiple topics
func (s *Pubsub) Sub(
	ctx context.Context,
	topics ...string,
) (<-chan interface{}, error) {
	ch := make(chan interface{})
	for _, t := range topics {
		err := s.makeSub(ctx, t, ch)
		if err != nil {
			return ch, err
		}
	}
	return ch, nil
}

func (s *Pubsub) makeSub(ctx context.Context, t string, ch chan<- interface{}) error {
	s.topics[t] = append(s.topics[t], subscriber{
		ctx: ctx,
		ch:  ch,
	})
	return nil
}

// Publish a message to the given topics
func (s *Pubsub) Publish(msg interface{}, topics ...string) error {
	for _, t := range topics {
		s.makePublish(msg, t)
	}
	return nil
}

func (s *Pubsub) makePublish(msg interface{}, t string) error {
	for i, sub := range s.topics[t] {
		select {
		case <-sub.ctx.Done():
			s.topics[t] = append(s.topics[t][:i], s.topics[t][:i+1]...)
		case sub.ch <- msg:
		}
	}
	return nil
}
