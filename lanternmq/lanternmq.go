package lanternmq

import (
	"github.com/streadway/amqp"
)

type MessageHandler func([]byte, *map[string]interface{}) error

type ChannelID string

type MessageQueue interface {
	Connect(username string, password string, host string, port string) error
	CreateChannel() (ChannelID, error)
	NumConcurrentMsgs(chID ChannelID, num int) error
	DeclareQueue(chID ChannelID, name string) error
	PublishToQueue(chID ChannelID, qName string, message string) error
	ConsumeFromQueue(chID ChannelID, qName string) (<-chan amqp.Delivery, error)
	ProcessMessages(msgs <-chan amqp.Delivery, handler MessageHandler, args *map[string]interface{}) error
	DeclareTarget(chID ChannelID, name string) error
	PublishToTarget(chID ChannelID, name string, routingKey string, message string) error
	DeclareTargetReceiveQueue(chID ChannelID, targetName string, qName string, routingKey string) error
	Close()
}
