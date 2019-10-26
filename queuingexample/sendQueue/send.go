package main

import (
	"log"
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
)

var mq lanternmq.MessageQueue

func failOnError(err error) {
	if err != nil {
		log.Fatalf("%s", err)
	}
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}

func main() {
	mq = &rabbitmq.MessageQueue{}

	conn, err := mq.Connect("guest", "guest", "localhost", "5672")
	failOnError(err)
	defer conn.Close()
	ch, err := mq.CreateChannel(conn)
	failOnError(err)
	defer ch.Close()

	err = mq.DeclareQueue(ch, "hello")
	failOnError(err)

	body := bodyFrom(os.Args)
	err = mq.PublishToQueue(ch, "hello", body)
	log.Printf(" [x] Sent %s", body)
	failOnError(err)
}
