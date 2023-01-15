package main

import (
	aws_helper "blRoute/aws-helper"
	"blRoute/processor"
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// reader is the main loop reading messages rom AWS queue
func reader(helper *aws_helper.AWSHelper, queueUrl string, maxMessages int64) {

	for {
		result, err := helper.GetAWSMessages(queueUrl, maxMessages, 20)
		if err != nil {
			log.Panicf("Critical error reading from aws queue: %v", err)
		}

		for _, msg := range result.Messages {
			b := msg.Body
			if b == nil || len(*b) == 0 {
				continue // empty message can be a ping
			}

			log.Printf("RCVD msg '%v'", *b)
		}
	}
}

func main() {
	var awsProfile = flag.String("profile", "default", "aws profile name")
	var awsRegion = flag.String("region", "eu-west-1", "aws region")
	var queueName = flag.String("qname", "bl-test-queue", "queue name")

	flag.Parse()
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: *awsProfile,
		Config: aws.Config{
			Region: awsRegion,
		},
	})

	if err != nil {
		log.Printf("Error initializing new session (profile %s region %s) : %v", *awsProfile, *awsRegion, err)
		os.Exit(1)
	}
	awsHelper := (&aws_helper.AWSHelper{}).Init(sess)
	qu, err := awsHelper.GetAWSQueueURL(*queueName)
	if err != nil {
		log.Printf("Error connecting to queue %s: %v", *queueName, err)
		os.Exit(1)
	}

	rand.Seed(time.Now().UnixNano())
	clientQName := fmt.Sprintf("bl-client-%09d-queue", rand.Int())
	createRes, err := awsHelper.CreateAWSQueue(clientQName) // create queue to receive messages
	if err != nil {
		fmt.Printf("Error creating queue %s: %v", clientQName, err)
		os.Exit(1)
	}
	defer awsHelper.DeleteAWSQueue(*createRes.QueueUrl)
	go reader(awsHelper, *createRes.QueueUrl, 1)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("[Press Q and then Enter to exit] Command ->\n")
		resp := ""
		if scanner.Scan() {
			resp = scanner.Text()
		}
		if len(resp) == 0 {
			continue
		}

		if strings.ToUpper(resp) == "Q" {
			fmt.Println("Good bye!")
			break
		}

		var bPayload []byte

		if resp[0] == '+' {
			sp := strings.Split(resp[1:], ":")
			if len(sp) > 1 {
				p := processor.NetPayload{
					OpCode: rune(resp[0]),
					Key:    sp[0],
					Value: &processor.QPayload{
						Body: sp[1],
					},
				}
				bPayload, _ = json.Marshal(p)
			}

		} else if resp[0] == '-' {
			p := processor.NetPayload{
				OpCode: rune(resp[0]),
				Key:    resp[1:],
			}
			bPayload, _ = json.Marshal(p)

		} else if resp[0] == '?' {
			p := processor.NetPayload{
				OpCode:      rune(resp[0]),
				Key:         resp[1:],
				ClientQueue: clientQName,
			}
			bPayload, _ = json.Marshal(p)

		} else {
			p := processor.NetPayload{
				OpCode:      rune(resp[0]),
				ClientQueue: clientQName,
			}
			bPayload, _ = json.Marshal(p)

		}
		if len(bPayload) > 0 {
			awsHelper.SendAWSMessage(*qu.QueueUrl, string(bPayload))
		}

	}

}
