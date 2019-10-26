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
	if (len(args) < 3) || os.Args[2] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[2:], " ")
	}
	return s
}

func severityFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "anonymous.info"
	} else {
		s = os.Args[1]
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

	err = mq.DeclareTarget(ch, "logs_topic")
	failOnError(err)

	body := bodyFrom(os.Args)
	severity := severityFrom(os.Args)
	err = mq.PublishToTarget(ch, "logs_topic", severity, body)
	failOnError(err)
	log.Printf(" [x] Sent %s", body)
}
