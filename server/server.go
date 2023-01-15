package main

import (
	aws_helper "blRoute/aws-helper"
	"blRoute/processor"
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

const DEBUG = true
const MAX_TIMOUT_SEC = 20

// reader is the main loop reading messages rom AWS queue
func reader(helper *aws_helper.AWSHelper, queueUrl string, maxMessages int64, proc processor.QProcessor) {

	for {
		result, err := helper.GetAWSMessages(queueUrl, maxMessages, MAX_TIMOUT_SEC)
		if err != nil {
			log.Panicf("Critical error reading from aws queue: %v", err)
		}

		for _, msg := range result.Messages {
			b := msg.Body
			if b == nil || len(*b) == 0 {
				continue // empty message can be a ping
			}

			if DEBUG {
				log.Printf("RCVD msg %v", *b)
			}
			proc.ProcPayload(*b)
		}
	}
}

// main
func main() {

	var awsProfile = flag.String("profile", "default", "aws profile name")
	var awsRegion = flag.String("region", "eu-west-1", "aws region")
	var queueName = flag.String("qname", "bl-test-queue", "queue name")
	var threads = flag.Int("threads", 3, "number of threads")

	flag.Parse()
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: *awsProfile,
		Config: aws.Config{
			Region: awsRegion,
		},
	})

	if err != nil {
		fmt.Printf("Error initializing new session (profile %s region %s) : %v", *awsProfile, *awsRegion, err)
		os.Exit(1)
	}

	awsHelper := (&aws_helper.AWSHelper{}).Init(sess)
	createRes, err := awsHelper.CreateAWSQueue(*queueName) // create queue to receive messages
	if err != nil {
		fmt.Printf("Error creating queue %s: %v", *queueName, err)
		os.Exit(1)
	}
	defer awsHelper.DeleteAWSQueue(*createRes.QueueUrl)

	log.Printf("Created a new queue with url: %s", *createRes.QueueUrl)

	var proc processor.QProcessor
	if *threads == 1 {
		proc = (&processor.QProcessorSt{}).Create(awsHelper)
	} else {
		proc = (&processor.QProcessorMt{}).Create(*threads, awsHelper) //create (multithreading) processor
	}

	go reader(awsHelper, *createRes.QueueUrl, 10, proc)

	log.Printf("Server is up and running on queue %s", *queueName)
	scanner := bufio.NewScanner(os.Stdin)
	resp := ""

	for {
		fmt.Printf("Press Enter to shut down..\n")
		if scanner.Scan() {
			resp = scanner.Text()
			if resp == "test" {

			}
		}
		return

	}
	fmt.Printf("\nServer is being shut down, have a nice day!\n")
}
