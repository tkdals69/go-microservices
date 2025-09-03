package adapters

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// AWSSQSBus implements message bus using AWS SQS
type AWSSQSBus struct {
	sqs *sqs.SQS
}

func NewAWSSQSBus(region string) (*AWSSQSBus, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		return nil, err
	}

	return &AWSSQSBus{
		sqs: sqs.New(sess),
	}, nil
}

func (b *AWSSQSBus) Publish(ctx context.Context, queueURL string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = b.sqs.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(data)),
	})

	return err
}

func (b *AWSSQSBus) Subscribe(ctx context.Context, queueURL string, handler func([]byte) error) error {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				result, err := b.sqs.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
					QueueUrl:            aws.String(queueURL),
					MaxNumberOfMessages: aws.Int64(1),
					WaitTimeSeconds:     aws.Int64(20), // Long polling
				})

				if err != nil {
					continue
				}

				for _, msg := range result.Messages {
					if err := handler([]byte(*msg.Body)); err == nil {
						// Delete message if processed successfully
						b.sqs.DeleteMessageWithContext(ctx, &sqs.DeleteMessageInput{
							QueueUrl:      aws.String(queueURL),
							ReceiptHandle: msg.ReceiptHandle,
						})
					}
				}
			}
		}
	}()

	return nil
}
