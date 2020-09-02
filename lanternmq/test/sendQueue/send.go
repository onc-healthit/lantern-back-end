// Sends a message to a queue. The first command line argument is the message string. To work with
// the accompanying `receive` executable, send strings composed of periods. Example: `go run send.go ...`
// The queue that the message is sent on is named 'hello'.
package main

import (
	"log"
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
	"github.com/onc-healthit/lantern-back-end/endpointmanager/sharedfunctions"
)

var mq lanternmq.MessageQueue

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
	defer mq.Close()

	err := mq.Connect("guest", "guest", "localhost", "5672")
	sharedfunctions.FailOnError("", err)
	ch, err := mq.CreateChannel()
	sharedfunctions.FailOnError("", err)

	err = mq.DeclareQueue(ch, "hello")
	sharedfunctions.FailOnError("", err)

	body := bodyFrom(os.Args)
	err = mq.PublishToQueue(ch, "hello", body)
	log.Printf(" [x] Sent %s", body)
	sharedfunctions.FailOnError("", err)
}
