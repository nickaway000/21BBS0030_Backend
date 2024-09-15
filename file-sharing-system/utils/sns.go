package utils

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/sns"
)

func PublishSNSMessage(subject, message string) error {
    sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-north-1")})
    if err != nil {
        return err
    }

    snsSvc := sns.New(sess)
    input := &sns.PublishInput{
        Message:  aws.String(message),
        Subject:  aws.String(subject),
        TopicArn: aws.String("arn:aws:sns:eu-north-1:051826734058:fileshare"), 
    }

    _, err = snsSvc.Publish(input)
    return err
}
