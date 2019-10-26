package mock

import (
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/streadway/amqp"
)

// Ensure MessageQueue implements lanternmq.MessageQueue.
var _ lanternmq.MessageQueue = &MessageQueue{}

type MessageQueue struct {
	ConnectFn      func(username string, password string, host string, port string) (*amqp.Connection, error)
	ConnectInvoked bool

	CreateChannelFn      func(conn *amqp.Connection) (*amqp.Channel, error)
	CreateChannelInvoked bool

	NumConcurrentMsgsFn      func(ch *amqp.Channel, num int) error
	NumConcurrentMsgsInvoked bool

	DeclareQueueFn      func(ch *amqp.Channel, name string) error
	DeclareQueueInvoked bool

	PublishToQueueFn      func(ch *amqp.Channel, qName string, message string) error
	PublishToQueueInvoked bool

	ConsumeFromQueueFn      func(ch *amqp.Channel, qName string) (<-chan amqp.Delivery, error)
	ConsumeFromQueueInvoked bool

	ProcessMessagesFn      func(msgs <-chan amqp.Delivery, handler lanternmq.MessageHandler, args *map[string]interface{}) error
	ProcessMessagesInvoked bool

	DeclareTargetFn      func(ch *amqp.Channel, name string) error
	DeclareTargetInvoked bool

	PublishToTargetFn      func(ch *amqp.Channel, name string, routingKey string, message string) error
	PublishToTargetInvoked bool

	DeclareTargetReceiveQueueFn      func(ch *amqp.Channel, targetName string, qName string, routingKey string) error
	DeclareTargetReceiveQueueInvoked bool
}

func (mq *MessageQueue) Connect(username string, password string, host string, port string) (*amqp.Connection, error) {
	mq.ConnectInvoked = true
	return mq.ConnectFn(username, password, host, port)
}

func (mq *MessageQueue) CreateChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	mq.CreateChannelInvoked = true
	return mq.CreateChannelFn(conn)
}

func (mq *MessageQueue) NumConcurrentMsgs(ch *amqp.Channel, num int) error {
	mq.NumConcurrentMsgsInvoked = true
	return mq.NumConcurrentMsgsFn(ch, num)
}

func (mq *MessageQueue) DeclareQueue(ch *amqp.Channel, name string) error {
	mq.DeclareQueueInvoked = true
	return mq.DeclareQueueFn(ch, name)
}

func (mq *MessageQueue) PublishToQueue(ch *amqp.Channel, qName string, message string) error {
	mq.PublishToQueueInvoked = true
	return mq.PublishToQueueFn(ch, qName, message)
}

func (mq *MessageQueue) ConsumeFromQueue(ch *amqp.Channel, qName string) (<-chan amqp.Delivery, error) {
	mq.ConsumeFromQueueInvoked = true
	return mq.ConsumeFromQueueFn(ch, qName)
}

func (mq *MessageQueue) ProcessMessages(msgs <-chan amqp.Delivery, handler lanternmq.MessageHandler, args *map[string]interface{}) error {
	mq.ProcessMessagesInvoked = true
	return mq.ProcessMessagesFn(msgs, handler, args)
}

func (mq *MessageQueue) DeclareTarget(ch *amqp.Channel, name string) error {
	mq.DeclareTargetInvoked = true
	return mq.DeclareTargetFn(ch, name)
}

func (mq *MessageQueue) PublishToTarget(ch *amqp.Channel, name string, routingKey string, message string) error {
	mq.PublishToTargetInvoked = true
	return mq.PublishToTargetFn(ch, name, routingKey, message)
}

func (mq *MessageQueue) DeclareTargetReceiveQueue(ch *amqp.Channel, targetName string, qName string, routingKey string) error {
	mq.DeclareTargetReceiveQueueInvoked = true
	return mq.DeclareTargetReceiveQueueFn(ch, targetName, qName, routingKey)
}
