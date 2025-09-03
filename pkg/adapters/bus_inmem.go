package adapters

import (
	"context"
	"encoding/json"
	"log"
)

// InMemoryBus implements a simple in-memory message bus for single-node operation
type InMemoryBus struct {
	channels map[string]chan []byte
}

func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{
		channels: make(map[string]chan []byte),
	}
}

func (b *InMemoryBus) CreateTopic(topicName string) error {
	if _, exists := b.channels[topicName]; !exists {
		b.channels[topicName] = make(chan []byte, 100) // buffered channel
	}
	return nil
}

func (b *InMemoryBus) Publish(ctx context.Context, topicName string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	ch, exists := b.channels[topicName]
	if !exists {
		if err := b.CreateTopic(topicName); err != nil {
			return err
		}
		ch = b.channels[topicName]
	}

	select {
	case ch <- data:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		log.Printf("Channel full, dropping message for topic: %s", topicName)
		return nil
	}
}

func (b *InMemoryBus) Subscribe(ctx context.Context, topicName string, handler func([]byte) error) error {
	ch, exists := b.channels[topicName]
	if !exists {
		if err := b.CreateTopic(topicName); err != nil {
			return err
		}
		ch = b.channels[topicName]
	}

	go func() {
		for {
			select {
			case msg := <-ch:
				if err := handler(msg); err != nil {
					log.Printf("Error handling message on topic %s: %v", topicName, err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}
