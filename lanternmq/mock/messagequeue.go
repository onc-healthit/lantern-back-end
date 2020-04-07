package mock

import (
	"github.com/onc-healthit/lantern-back-end/lanternmq"
)

// Ensure MessageQueue implements lanternmq.MessageQueue.
var _ lanternmq.MessageQueue = &MessageQueue{}

// MessageQueue implements the lanternmq.MessageQueue interface and allows mock implementations of a MessageQueue.
// Each MessageQueue method calls the corresponding method <methodName>Fn as assigned in the mock MessageQueue
// structure. It also assigns <methodName>Invoked to true when <methodName> is called.
type MessageQueue struct {
	ConnectFn func(username string, password string, host string, port string) error

	CreateChannelFn func() (lanternmq.ChannelID, error)

	NumConcurrentMsgsFn func(chID lanternmq.ChannelID, num int) error

	DeclareQueueFn func(chID lanternmq.ChannelID, name string) error

	PublishToQueueFn func(chID lanternmq.ChannelID, qName string, message string) error

	ConsumeFromQueueFn func(chID lanternmq.ChannelID, qName string) (lanternmq.Messages, error)

	ProcessMessagesFn func(msgs lanternmq.Messages, handler lanternmq.MessageHandler, args *map[string]interface{}, errs chan<- error)

	DeclareExchangeFn func(chID lanternmq.ChannelID, name string, exchangeType string) error

	PublishToExchangeFn func(chID lanternmq.ChannelID, name string, routingKey string, message string) error

	DeclareExchangeReceiveQueueFn func(chID lanternmq.ChannelID, topicName string, qName string, routingKey string) error

	CloseFn func()
}

// Connect mocks lanternmq.Connect and calls mq.ConnectFn with the given arguments.
func (mq *MessageQueue) Connect(username string, password string, host string, port string) error {
	return mq.ConnectFn(username, password, host, port)
}

// CreateChannel mocks lanternmq.CreateChannel and calls mq.CreateChannelFn with the given arguments.
func (mq *MessageQueue) CreateChannel() (lanternmq.ChannelID, error) {
	return mq.CreateChannelFn()
}

// NumConcurrentMsgs mocks lanternmq.NumConcurrentMsgs and calls mq.NumConcurrentMsgsFn with the given arguments.
func (mq *MessageQueue) NumConcurrentMsgs(chID lanternmq.ChannelID, num int) error {
	return mq.NumConcurrentMsgsFn(chID, num)
}

// DeclareQueue mocks lanternmq.DeclareQueue and calls mq.DeclareQueueFn with the given arguments.
func (mq *MessageQueue) DeclareQueue(chID lanternmq.ChannelID, name string) error {
	return mq.DeclareQueueFn(chID, name)
}

// PublishToQueue mocks lanternmq.PublishToQueue and calls mq.PublishToQueueFn with the given arguments.
func (mq *MessageQueue) PublishToQueue(chID lanternmq.ChannelID, qName string, message string) error {
	return mq.PublishToQueueFn(chID, qName, message)
}

// ConsumeFromQueue mocks lanternmq.ConsumeFromQueue and calls mq.ConsumeFromQueueFn with the given arguments.
func (mq *MessageQueue) ConsumeFromQueue(chID lanternmq.ChannelID, qName string) (lanternmq.Messages, error) {
	return mq.ConsumeFromQueueFn(chID, qName)
}

// ProcessMessages mocks lanternmq.ProcessMessages and calls mq.ProcessMessagesFn with the given arguments.
func (mq *MessageQueue) ProcessMessages(msgs lanternmq.Messages, handler lanternmq.MessageHandler, args *map[string]interface{}, errs chan<- error) {
	mq.ProcessMessagesFn(msgs, handler, args, errs)
}

// DeclareExchange mocks lanternmq.DeclareExchange and calls mq.DeclareExchangeFn with the given arguments.
func (mq *MessageQueue) DeclareExchange(chID lanternmq.ChannelID, name string, exchangeType string) error {
	return mq.DeclareExchangeFn(chID, name, exchangeType)
}

// PublishToExchange mocks lanternmq.PublishToExchange and calls mq.PublishToExchangeFn with the given arguments.
func (mq *MessageQueue) PublishToExchange(chID lanternmq.ChannelID, name string, routingKey string, message string) error {
	return mq.PublishToExchangeFn(chID, name, routingKey, message)
}

// DeclareExchangeReceiveQueue mocks lanternmq.DeclareExchangeReceiveQueue and calls mq.DeclareExchangeReceiveQueueFn with the given arguments.
func (mq *MessageQueue) DeclareExchangeReceiveQueue(chID lanternmq.ChannelID, topicName string, qName string, routingKey string) error {
	return mq.DeclareExchangeReceiveQueueFn(chID, topicName, qName, routingKey)
}

// Close mocks lanternmq.Close and calls mq.CloseFn with the given arguments.
func (mq *MessageQueue) Close() {
	mq.CloseFn()
}
