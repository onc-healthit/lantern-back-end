package main

import (
	"log"
	"os"
	"strings"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
)

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
	conn, err := lanternmq.Connect("guest", "guest", "localhost", "5672")
	failOnError(err)
	defer conn.Close()
	ch, err := lanternmq.CreateChannel(conn)
	failOnError(err)
	defer ch.Close()

	q, err := lanternmq.CreateQueue(ch, "hello")
	failOnError(err)

	body := bodyFrom(os.Args)
	err = lanternmq.PublishToQueue(ch, q, body)
	log.Printf(" [x] Sent %s", body)
	failOnError(err)
}
