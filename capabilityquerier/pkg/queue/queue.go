package queue

import (
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/rabbitmq"
	"github.com/pkg/errors"
)

// ConnectToQueue creates a connection to a queue at the given location with the given credentials.
func ConnectToQueue(qUser, qPassword, qHost, qPort, qName string) (lanternmq.MessageQueue, lanternmq.ChannelID, error) {
	mq := &rabbitmq.MessageQueue{}
	err := mq.Connect(qUser, qPassword, qHost, qPort)
	if err != nil {
		return nil, nil, err
	}
	ch, err := mq.CreateChannel()
	if err != nil {
		return nil, nil, err
	}
	exists, err := mq.QueueExists(ch, qName)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, errors.Errorf("queue %s does not exist", qName)
	}

	return mq, ch, nil
}
