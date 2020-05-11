package accessqueue

import (
	"context"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/pkg/errors"
)

// ConnectToServerAndQueue creates a connection to an exchange at the given location with the given credentials.
// then connects to the queue with the given queue name
func ConnectToServerAndQueue(qUser, qPassword, qHost, qPort, qName string) (lanternmq.MessageQueue, lanternmq.ChannelID, error) {
	mq := &lanternmq.MessageQueue{}
	err := mq.Connect(qUser, qPassword, qHost, qPort)
	if err != nil {
		return nil, nil, err
	}
	ch, err := mq.CreateChannel()
	if err != nil {
		return nil, nil, err
	}
	return ConnectToQueue(mq, ch, qName)
}

// ConnectToQueue uses the given connection to connect to the queue with the given queue name
func ConnectToQueue(mq lanternmq.MessageQueue, ch lanternmq.ChannelID, qName string) (lanternmq.MessageQueue, lanternmq.ChannelID, error) {
	exists, err := mq.QueueExists(ch, qName)
	if err != nil {
		return nil, nil, err
	}
	if !exists {
		return nil, nil, errors.Errorf("queue %s does not exist", qName)
	}

	return mq, ch, nil
}

// SendToQueue publishes a message to the given queue
func SendToQueue(
	ctx context.Context,
	message string,
	mq *lanternmq.MessageQueue,
	ch *lanternmq.ChannelID,
	queueName string) error {

	// don't send the message if the context is done
	select {
	case <-ctx.Done():
		return errors.Wrap(ctx.Err(), "unable to send message to queue - context ended")
	default:
		// ok
	}

	err := (*mq).PublishToQueue(*ch, queueName, message)
	if err != nil {
		return err
	}

	return nil
}
