package main

import (
	"bytes"
	"log"
	"os"
	"time"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
)

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
	conn, err := lanternmq.Connect("guest", "guest", "localhost", "5672")
	failOnError(err)
	defer conn.Close()
	ch, err := lanternmq.CreateChannel(conn)
	failOnError(err)
	defer ch.Close()

	lanternmq.NumConcurrentMsgs(ch, 1)

	// Queue
	err = lanternmq.CreateQueue(ch, "hello")
	failOnError(err)
	msgs, err := lanternmq.ConsumeFromQueue(ch, "hello")
	failOnError(err)

	// Topic
	tqName := os.Args[1]
	err = lanternmq.DeclareTarget(ch, "logs_topic")
	failOnError(err)
	err = lanternmq.CreateTargetReceiveQueue(ch, "logs_topic", tqName, "warning")
	failOnError(err)
	err = lanternmq.CreateTargetReceiveQueue(ch, "logs_topic", tqName, "error")
	failOnError(err)
	tmsgs, err := lanternmq.ConsumeFromQueue(ch, tqName)
	failOnError(err)

	forever := make(chan bool)

	go lanternmq.ProcessMessages(msgs, handleQueueMessage, nil)
	go lanternmq.ProcessMessages(tmsgs, handleTargetMessage, nil)

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
