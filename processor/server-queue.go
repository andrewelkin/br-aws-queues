package processor

import (
	aws_helper "blRoute/aws-helper"
	"encoding/json"
	"fmt"
	wPool "github.com/gammazero/workerpool"
	"log"
)

type QPayload struct {
	Body string `json:"body"`
}

type NetPayload struct {
	OpCode      rune      `json:"op"`
	Key         string    `json:"key"`
	Value       *QPayload `json:"value"`
	ClientQueue string    `json:"ret_queue"`
}

type QProcessor interface {
	ProcPayload(string)
}

const DEBUG = true

// -----------------------------------------------------------------------------------------[AE: 2023-01-13]-----------/

// QProcessorSt single-thread processor for payloads
type QProcessorSt struct {
	queue     *OrderedMap
	awsHelper *aws_helper.AWSHelper
}

// ProcPayload process payload interface method
// --> Input:
// b     string     payload (json string)
func (q *QProcessorSt) ProcPayload(b string) {

	var p NetPayload

	if err := json.Unmarshal([]byte(b), &p); err != nil {
		log.Printf("Error unmarshalling message (ignoring): %v", b)
		return
	}
	switch p.OpCode {
	case '+': // add item
		q.queue.Add(p.Key, p.Value)
	case '-': // delete item
		q.queue.Delete(p.Key)

	case '?': // get one item
		i, ok := q.queue.Get(p.Key)
		if ok {
			q.awsHelper.SendAWSMessageToQName(p.ClientQueue, fmt.Sprintf("%s -> %s", p.Key, i.Body))
		} else {
			q.awsHelper.SendAWSMessageToQName(p.ClientQueue, fmt.Sprintf("Key %s not found", p.Key))
		}

	case '*': // list all items
		keys, items := q.queue.FlattenWithKeys()
		message := ""
		for j, k := range keys {
			message += fmt.Sprintf("\n%s -> %s", k, items[j].Body)
		}
		q.awsHelper.SendAWSMessageToQName(p.ClientQueue, message)

	case 'd': // debug
		keys := q.queue.GetKeys()
		log.Printf("Current queue state: %v", keys)
	}

	if DEBUG {
		keys := q.queue.GetKeys()
		log.Printf("Current queue state: %v", keys)
	}
}

// Create creates single-thread processor for payloads
// --> Input:
// helper       *aws_helper.AWSHelper     aws helper
// <-- Output:
// 1) QProcessor     processor interface
func (q *QProcessorSt) Create(helper *aws_helper.AWSHelper) QProcessor {
	q.queue = NewOrderedMap()
	q.awsHelper = helper
	return q
}

// -----------------------------------------------------------------------------------------[AE: 2023-01-13]-----------/

// QProcessorMt multi-thread processor for payloads
type QProcessorMt struct {
	queue      *OrderedMap
	addCh      chan NetPayload
	delCh      chan string
	getCh      chan NetPayload
	getAllCh   chan NetPayload
	workerPool *wPool.WorkerPool
	awsHelper  *aws_helper.AWSHelper
}

// main loop to operate on internal queue
func (q *QProcessorMt) process() {
	for {
		select {
		case p := <-q.addCh:
			q.queue.Add(p.Key, p.Value)
			q.debug()
		case k := <-q.delCh:
			// delete key k
			q.queue.Delete(k)
			q.debug()

		case p := <-q.getCh:
			i, ok := q.queue.Get(p.Key)
			if ok {
				q.awsHelper.SendAWSMessageToQName(p.ClientQueue, fmt.Sprintf("%s -> %s", p.Key, i.Body))
			} else {
				q.awsHelper.SendAWSMessageToQName(p.ClientQueue, fmt.Sprintf("Key %s not found", p.Key))
			}
		case p := <-q.getAllCh:
			keys, items := q.queue.FlattenWithKeys()
			message := ""
			for j, k := range keys {
				message += fmt.Sprintf("\n%s -> %s", k, items[j].Body)
			}
			q.awsHelper.SendAWSMessageToQName(p.ClientQueue, message)
		}
	}
}

// debug prints current state of the internal queue
func (q *QProcessorMt) debug() {
	if DEBUG {
		keys := q.queue.GetKeys()
		log.Printf("Current queue state: %v", keys)
	}
}

// Create creates multi-thread processor for payloads
// --> Input:
// poolSize     int                       number of goroutines to process payloads
// helper       *aws_helper.AWSHelper     aws helper
// <-- Output:
// 1) QProcessor     processor interface
func (q *QProcessorMt) Create(poolSize int, helper *aws_helper.AWSHelper) QProcessor {
	q.addCh = make(chan NetPayload, 1024)
	q.delCh = make(chan string, 1024)
	q.getAllCh = make(chan NetPayload, 1024)
	q.getCh = make(chan NetPayload, 1024)
	q.queue = NewOrderedMap()
	q.workerPool = wPool.New(poolSize) // init the worker pool
	q.awsHelper = helper
	go q.process()
	return q
}

// ProcPayload process payload interface method
// --> Input:
// b     string     payload (json string)
func (q *QProcessorMt) ProcPayload(b string) {

	f := func() {
		var p NetPayload
		if err := json.Unmarshal([]byte(b), &p); err != nil {
			log.Printf("Error unmarshalling message (ignoring): %v", b)
		}

		// here we might do some heavy processing of the payload
		switch p.OpCode {
		case '+': // add item
			q.addCh <- p
		case '-': // delete item
			q.delCh <- p.Key
		case '?': // get one item
			q.getCh <- p
		case '*': // list all items
			q.getAllCh <- p

		case 'd': // debug
			keys := q.queue.GetKeys()
			log.Printf("Current queue state: %v", keys)
		}

	}
	q.workerPool.Submit(f)
}
