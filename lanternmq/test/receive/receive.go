// Receives messages from the queue named 'hello' and from the topic with target name 'logs_topic'
// and topics 'error' and 'warning'.
// When receiving messages from the queue, it counts how many periods are in the message and waits that
// main seconds. For example, it will wait 3 seconds after receiving the message '...'. This is
// supposed to simulate work being performed.
package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"time"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/sharedfunctions"
)

var mq lanternmq.MessageQueue

func handleQueueMessage(msg []byte, _ *map[string]interface{}) error {
	log.Printf("QUEUE: Received a message: %s", msg)
	dotCount := bytes.Count(msg, []byte("."))
	t := time.Duration(dotCount)
	time.Sleep(t * time.Second)
	log.Printf("Done")
	return nil
}

func handleTopicMessage(msg []byte, _ *map[string]interface{}) error {
	log.Printf("TOPIC: Received message: %s\n", msg)
	return nil
}

func main() {
	mq = &rabbitmq.MessageQueue{}
	defer mq.Close()

	err := mq.Connect("guest", "guest", "localhost", "5672")
	sharedfunctions.FailOnError("", err)
	ch, err := mq.CreateChannel()
	sharedfunctions.FailOnError("", err)

	err = mq.NumConcurrentMsgs(ch, 1)
	sharedfunctions.FailOnError("", err)

	// Queue
	err = mq.DeclareQueue(ch, "hello")
	sharedfunctions.FailOnError("", err)
	msgs, err := mq.ConsumeFromQueue(ch, "hello")
	sharedfunctions.FailOnError("", err)

	// Topic
	tqName := os.Args[1]
	err = mq.DeclareExchange(ch, "logs_topic", "topic")
	sharedfunctions.FailOnError("", err)
	err = mq.DeclareExchangeReceiveQueue(ch, "logs_topic", tqName, "warning")
	sharedfunctions.FailOnError("", err)
	err = mq.DeclareExchangeReceiveQueue(ch, "logs_topic", tqName, "error")
	sharedfunctions.FailOnError("", err)
	tmsgs, err := mq.ConsumeFromQueue(ch, tqName)
	sharedfunctions.FailOnError("", err)

	forever := make(chan bool)

	errs := make(chan error)
	defer close(errs)

	ctx := context.Background()

	go mq.ProcessMessages(ctx, msgs, handleQueueMessage, nil, errs)
	go mq.ProcessMessages(ctx, tmsgs, handleTopicMessage, nil, errs)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")

	for err = range errs {
		panic(err.Error())
	}
	<-forever
}
