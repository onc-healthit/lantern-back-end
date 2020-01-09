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

	DeclareTopicFn func(chID lanternmq.ChannelID, name string) error

	PublishToTopicFn func(chID lanternmq.ChannelID, name string, routingKey string, message string) error

	DeclareTopicReceiveQueueFn func(chID lanternmq.ChannelID, topicName string, qName string, routingKey string) error

	CloseFn func()
}

// Connect mocks lanternmq.Connect and sets mq.ConnectInvoked to true and calls mq.ConnectFn with the given arguments.
func (mq *MessageQueue) Connect(username string, password string, host string, port string) error {
	return mq.ConnectFn(username, password, host, port)
}

// CreateChannel mocks lanternmq.CreateChannel and sets mq.CreateChannelInvoked to true and calls mq.CreateChannelFn with the given arguments.
func (mq *MessageQueue) CreateChannel() (lanternmq.ChannelID, error) {
	return mq.CreateChannelFn()
}

// NumConcurrentMsgs mocks lanternmq.NumConcurrentMsgs and sets mq.NumConcurrentMsgsInvoked to true and calls mq.NumConcurrentMsgsFn with the given arguments.
func (mq *MessageQueue) NumConcurrentMsgs(chID lanternmq.ChannelID, num int) error {
	return mq.NumConcurrentMsgsFn(chID, num)
}

// DeclareQueue mocks lanternmq.DeclareQueue and sets mq.DeclareQueueInvoked to true and calls mq.DeclareQueueFn with the given arguments.
func (mq *MessageQueue) DeclareQueue(chID lanternmq.ChannelID, name string) error {
	return mq.DeclareQueueFn(chID, name)
}

// PublishToQueue mocks lanternmq.PublishToQueue and sets mq.PublishToQueueInvoked to true and calls mq.PublishToQueueFn with the given arguments.
func (mq *MessageQueue) PublishToQueue(chID lanternmq.ChannelID, qName string, message string) error {
	return mq.PublishToQueueFn(chID, qName, message)
}

// ConsumeFromQueue mocks lanternmq.ConsumeFromQueue and sets mq.ConsumeFromQueueInvoked to true and calls mq.ConsumeFromQueueFn with the given arguments.
func (mq *MessageQueue) ConsumeFromQueue(chID lanternmq.ChannelID, qName string) (lanternmq.Messages, error) {
	return mq.ConsumeFromQueueFn(chID, qName)
}

// ProcessMessages mocks lanternmq.ProcessMessages and sets mq.ProcessMessagesInvoked to true and calls mq.ProcessMessagesFn with the given arguments.
func (mq *MessageQueue) ProcessMessages(msgs lanternmq.Messages, handler lanternmq.MessageHandler, args *map[string]interface{}, errs chan<- error) {
	mq.ProcessMessagesFn(msgs, handler, args, errs)
}

// DeclareTopic mocks lanternmq.DeclareTopic and sets mq.DeclareTopicInvoked to true and calls mq.DeclareTopicFn with the given arguments.
func (mq *MessageQueue) DeclareTopic(chID lanternmq.ChannelID, name string) error {
	return mq.DeclareTopicFn(chID, name)
}

// PublishToTopic mocks lanternmq.PublishToTopic and sets mq.PublishToTopicInvoked to true and calls mq.PublishToTopicFn with the given arguments.
func (mq *MessageQueue) PublishToTopic(chID lanternmq.ChannelID, name string, routingKey string, message string) error {
	return mq.PublishToTopicFn(chID, name, routingKey, message)
}

// DeclareTopicReceiveQueue mocks lanternmq.DeclareTopicReceiveQueue and sets mq.DeclareTopicReceiveQueueInvoked to true and calls mq.DeclareTopicReceiveQueueFn with the given arguments.
func (mq *MessageQueue) DeclareTopicReceiveQueue(chID lanternmq.ChannelID, topicName string, qName string, routingKey string) error {
	return mq.DeclareTopicReceiveQueueFn(chID, topicName, qName, routingKey)
}

// Close mocks lanternmq.Close and sets mq.CloseInvoked to true and calls mq.CloseFn with the given arguments.
func (mq *MessageQueue) Close() {
	mq.CloseFn()
}
