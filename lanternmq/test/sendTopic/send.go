// Sends a message as a queue topic. The first command line argument is the topic string.
// The second command line argument is the message string. If the topic string is missing, "anonymous.info"
// is used as the topic. If the message string is missing, "hello" is used as the message.
// The topic messages are posted to the target "logs_topic".
package main

import (
	"log"
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/helpers"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
)

var mq lanternmq.MessageQueue

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
	defer mq.Close()

	err := mq.Connect("guest", "guest", "localhost", "5672")
	helpers.FailOnError("", err)
	ch, err := mq.CreateChannel()
	helpers.FailOnError("", err)

	err = mq.DeclareExchange(ch, "logs_topic", "topic")
	helpers.FailOnError("", err)

	body := bodyFrom(os.Args)
	severity := severityFrom(os.Args)
	err = mq.PublishToExchange(ch, "logs_topic", severity, body)
	helpers.FailOnError("", err)
	log.Printf(" [x] Sent %s", body)
}
