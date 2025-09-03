package adapters

import (
	"context"
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azservicebus"
)

// AzureServiceBus implements message bus using Azure Service Bus
type AzureServiceBus struct {
	client *azservicebus.Client
}

func NewAzureServiceBus(connectionString string) (*AzureServiceBus, error) {
	client, err := azservicebus.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, err
	}

	return &AzureServiceBus{
		client: client,
	}, nil
}

func (b *AzureServiceBus) Publish(ctx context.Context, topicName string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	sender, err := b.client.NewSender(topicName, nil)
	if err != nil {
		return err
	}
	defer sender.Close(ctx)

	msg := &azservicebus.Message{
		Body: data,
	}

	return sender.SendMessage(ctx, msg, nil)
}

func (b *AzureServiceBus) Subscribe(ctx context.Context, topicName, subscriptionName string, handler func([]byte) error) error {
	receiver, err := b.client.NewReceiverForSubscription(topicName, subscriptionName, nil)
	if err != nil {
		return err
	}
	defer receiver.Close(ctx)

	go func() {
		for {
			messages, err := receiver.ReceiveMessages(ctx, 1, nil)
			if err != nil {
				return
			}

			for _, msg := range messages {
				if err := handler(msg.Body); err == nil {
					receiver.CompleteMessage(ctx, msg, nil)
				} else {
					receiver.AbandonMessage(ctx, msg, nil)
				}
			}
		}
	}()

	return nil
}

func (b *AzureServiceBus) Close() error {
	return b.client.Close(context.Background())
}
