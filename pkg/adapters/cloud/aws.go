package cloud

import (
    "context"
    "fmt"
    "github.com/tkdals69/go-microservicesaws/aws-sdk-go/aws"
    "github.com/tkdals69/go-microservicesaws/aws-sdk-go/aws/session"
    "github.com/tkdals69/go-microservicesaws/aws-sdk-go/service/s3"
)

type AWSAdapter struct {
    session *session.Session
    s3      *s3.S3
}

func NewAWSAdapter(region string) (*AWSAdapter, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(region),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create AWS session: %w", err)
    }

    return &AWSAdapter{
        session: sess,
        s3:      s3.New(sess),
    }, nil
}

func (a *AWSAdapter) UploadToS3(bucket, key string, body []byte) error {
    _, err := a.s3.PutObject(&s3.PutObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
        Body:   aws.ReadSeekCloser(bytes.NewReader(body)),
    })
    if err != nil {
        return fmt.Errorf("failed to upload to S3: %w", err)
    }
    return nil
}

func (a *AWSAdapter) GetFromS3(bucket, key string) ([]byte, error) {
    result, err := a.s3.GetObject(&s3.GetObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get object from S3: %w", err)
    }
    defer result.Body.Close()

    body, err := io.ReadAll(result.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read object body: %w", err)
    }

    return body, nil
}