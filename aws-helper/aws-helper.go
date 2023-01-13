package aws_helper

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type AWSHelper struct {
	sqsClient *sqs.SQS
}

func (h *AWSHelper) Init(sess *session.Session) *AWSHelper {
	h.sqsClient = sqs.New(sess)
	return h
}

// DeleteAWSQueue deletes message from the AWS queue
func (h *AWSHelper) DeleteAWSQueue(queueURL string) error {

	_, err := h.sqsClient.DeleteQueue(&sqs.DeleteQueueInput{
		QueueUrl: &queueURL,
	})
	return err
}

// CreateAWSQueue creates queue
func (h *AWSHelper) CreateAWSQueue(queueName string) (*sqs.CreateQueueOutput, error) {

	result, err := h.sqsClient.CreateQueue(&sqs.CreateQueueInput{
		QueueName: &queueName,
		Attributes: map[string]*string{
			"DelaySeconds":      aws.String("0"),
			"VisibilityTimeout": aws.String("20"),
		},
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// SendAWSMessageToQName sends message using queue name
func (h *AWSHelper) SendAWSMessageToQName(queueName string, messageBody string) error {

	r, err := h.GetAWSQueueURL(queueName)
	if err != nil {
		return err
	}
	return h.SendAWSMessage(*r.QueueUrl, messageBody)
}

// GetAWSQueueURL gets queue URL by name
func (h *AWSHelper) GetAWSQueueURL(queue string) (*sqs.GetQueueUrlOutput, error) {

	result, err := h.sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queue,
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// SendAWSMessage sends message using queue url
func (h *AWSHelper) SendAWSMessage(queueUrl string, messageBody string) error {

	_, err := h.sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: aws.String(messageBody),
	})

	return err
}

// GetAWSMessages receives one or more messages from the queue
func (h *AWSHelper) GetAWSMessages(queueUrl string, maxMessages int64) (*sqs.ReceiveMessageOutput, error) {

	msgResult, err := h.sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: aws.Int64(maxMessages),
	})

	if err != nil {
		return nil, err
	}
	for _, m := range msgResult.Messages {
		h.sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      &queueUrl,
			ReceiptHandle: m.ReceiptHandle,
		})
	}

	return msgResult, nil
}
