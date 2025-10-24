package eventbus

import "context"

type EventBus interface {
	Publish(ctx context.Context, topic string, evt *Event) error
	Subscribe(ctx context.Context, topic, group string, handler EventHandler) error
	Close() error
}

type EventHandler func(ctx context.Context, evt *Event) error
