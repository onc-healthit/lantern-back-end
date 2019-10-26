package mock

import (
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/streadway/amqp"
)

// Ensure MessageQueue implements lanternmq.MessageQueue.
var _ lanternmq.MessageQueue = &MessageQueue{}

type MessageQueue struct {
	ConnectFn      func(username string, password string, host string, port string) error
	ConnectInvoked bool

	CreateChannelFn      func() (lanternmq.ChannelID, error)
	CreateChannelInvoked bool

	NumConcurrentMsgsFn      func(chID lanternmq.ChannelID, num int) error
	NumConcurrentMsgsInvoked bool

	DeclareQueueFn      func(chID lanternmq.ChannelID, name string) error
	DeclareQueueInvoked bool

	PublishToQueueFn      func(chID lanternmq.ChannelID, qName string, message string) error
	PublishToQueueInvoked bool

	ConsumeFromQueueFn      func(chID lanternmq.ChannelID, qName string) (<-chan amqp.Delivery, error)
	ConsumeFromQueueInvoked bool

	ProcessMessagesFn      func(msgs <-chan amqp.Delivery, handler lanternmq.MessageHandler, args *map[string]interface{}) error
	ProcessMessagesInvoked bool

	DeclareTargetFn      func(chID lanternmq.ChannelID, name string) error
	DeclareTargetInvoked bool

	PublishToTargetFn      func(chID lanternmq.ChannelID, name string, routingKey string, message string) error
	PublishToTargetInvoked bool

	DeclareTargetReceiveQueueFn      func(chID lanternmq.ChannelID, targetName string, qName string, routingKey string) error
	DeclareTargetReceiveQueueInvoked bool

	CloseFn      func()
	CloseInvoked bool
}

func (mq *MessageQueue) Connect(username string, password string, host string, port string) error {
	mq.ConnectInvoked = true
	return mq.ConnectFn(username, password, host, port)
}

func (mq *MessageQueue) CreateChannel() (lanternmq.ChannelID, error) {
	mq.CreateChannelInvoked = true
	return mq.CreateChannelFn(conn)
}

func (mq *MessageQueue) NumConcurrentMsgs(chID lanternmq.ChannelID, num int) error {
	mq.NumConcurrentMsgsInvoked = true
	return mq.NumConcurrentMsgsFn(chID, num)
}

func (mq *MessageQueue) DeclareQueue(chID lanternmq.ChannelID, name string) error {
	mq.DeclareQueueInvoked = true
	return mq.DeclareQueueFn(chID, name)
}

func (mq *MessageQueue) PublishToQueue(chID lanternmq.ChannelID, qName string, message string) error {
	mq.PublishToQueueInvoked = true
	return mq.PublishToQueueFn(chID, qName, message)
}

func (mq *MessageQueue) ConsumeFromQueue(chID lanternmq.ChannelID, qName string) (<-chan amqp.Delivery, error) {
	mq.ConsumeFromQueueInvoked = true
	return mq.ConsumeFromQueueFn(chID, qName)
}

func (mq *MessageQueue) ProcessMessages(msgs <-chan amqp.Delivery, handler lanternmq.MessageHandler, args *map[string]interface{}) error {
	mq.ProcessMessagesInvoked = true
	return mq.ProcessMessagesFn(msgs, handler, args)
}

func (mq *MessageQueue) DeclareTarget(chID lanternmq.ChannelID, name string) error {
	mq.DeclareTargetInvoked = true
	return mq.DeclareTargetFn(chID, name)
}

func (mq *MessageQueue) PublishToTarget(chID lanternmq.ChannelID, name string, routingKey string, message string) error {
	mq.PublishToTargetInvoked = true
	return mq.PublishToTargetFn(chID, name, routingKey, message)
}

func (mq *MessageQueue) DeclareTargetReceiveQueue(chID lanternmq.ChannelID, targetName string, qName string, routingKey string) error {
	mq.DeclareTargetReceiveQueueInvoked = true
	return mq.DeclareTargetReceiveQueueFn(chID, targetName, qName, routingKey)
}

func (mq *MessageQueue) Close() {
	mq.CloseInvoked = true
	mq.CloseFn()
}
