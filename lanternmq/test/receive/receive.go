package main

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
)

var mq lanternmq.MessageQueue

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func handleQueueMessage(msg []byte, _ *map[string]interface{}) error {
	log.Printf("QUEUE: Received a message: %s", msg)
	dotCount := bytes.Count(msg, []byte("."))
	t := time.Duration(dotCount)
	time.Sleep(t * time.Second)
	log.Printf("Done")
	return nil
}

func handleTargetMessage(msg []byte, _ *map[string]interface{}) error {
	log.Printf("TARGET: Received message: %s\n", string(msg))
	return nil
}

func main() {
	mq = &rabbitmq.MessageQueue{}
	defer mq.Close()

	err := mq.Connect("guest", "guest", "localhost", "5672")
	failOnError(err)
	ch, err := mq.CreateChannel()
	failOnError(err)

	mq.NumConcurrentMsgs(ch, 1)

	// Queue
	err = mq.DeclareQueue(ch, "hello")
	failOnError(err)
	msgs, err := mq.ConsumeFromQueue(ch, "hello")
	failOnError(err)

	// Topic
	tqName := os.Args[1]
	err = mq.DeclareTarget(ch, "logs_topic")
	failOnError(err)
	err = mq.DeclareTargetReceiveQueue(ch, "logs_topic", tqName, "warning")
	failOnError(err)
	err = mq.DeclareTargetReceiveQueue(ch, "logs_topic", tqName, "error")
	failOnError(err)
	tmsgs, err := mq.ConsumeFromQueue(ch, tqName)
	failOnError(err)

	forever := make(chan bool)

	go mq.ProcessMessages(msgs, handleQueueMessage, nil)
	go mq.ProcessMessages(tmsgs, handleTargetMessage, nil)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
