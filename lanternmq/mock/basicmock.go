package mock

import (
	"errors"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
)

// BasicMockMessageQueue is a basic implementation of a queue for byte arrays. It only sends and receives
// on a single queue and ignores things like channel and queue name.
// It can only handle 20 messages on the queue at a time.
type BasicMockMessageQueue struct {
	Queue chan []byte
	MessageQueue
}

// NewBasicMockMessageQueue initializes a BasicMockMessageQueue.
// Currently does not initialize any topic related functions.
func NewBasicMockMessageQueue() lanternmq.MessageQueue {
	mq := BasicMockMessageQueue{}
	mq.Queue = make(chan []byte, 20)

	mq.ConnectFn = func(username string, password string, host string, port string) error {
		return nil
	}

	mq.CreateChannelFn = func() (lanternmq.ChannelID, error) {
		return 1, nil
	}

	mq.NumConcurrentMsgsFn = func(chID lanternmq.ChannelID, num int) error {
		return nil
	}

	mq.DeclareQueueFn = func(chID lanternmq.ChannelID, name string) error {
		return nil
	}

	mq.PublishToQueueFn = func(chID lanternmq.ChannelID, qName string, message string) error {
		if len(mq.Queue) < 20 {
			mq.Queue <- []byte(message)
		} else {
			return errors.New("queue full - unable to add new message")
		}
		return nil
	}

	mq.ConsumeFromQueueFn = func(chID lanternmq.ChannelID, qName string) (lanternmq.Messages, error) {
		return nil, nil
	}

	mq.ProcessMessagesFn = func(msgs lanternmq.Messages, handler lanternmq.MessageHandler, args *map[string]interface{}, errs chan<- error) {
		for msg := range mq.Queue {
			handler(msg, args)
		}
	}

	mq.CloseFn = func() {}
	return &mq
}
