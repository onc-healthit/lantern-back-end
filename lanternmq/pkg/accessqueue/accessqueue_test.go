package accessqueue

import (
	"context"
	"testing"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/mock"
	"github.com/pkg/errors"
)

func Test_SendToQueue(t *testing.T) {
	var ch lanternmq.ChannelID
	var ctx context.Context
	var err error

	message := "this is a message"
	mq := mock.NewBasicMockMessageQueue()
	ch = 1
	queueName := "queue name"

	// basic test

	ctx = context.Background()

	err = SendToQueue(ctx, message, &mq, &ch, queueName)
	th.Assert(t, err == nil, err)

	th.Assert(t, len(mq.(*mock.BasicMockMessageQueue).Queue) == 1, "expected a message to be in the queue")

	bRcvMsg := <-mq.(*mock.BasicMockMessageQueue).Queue
	rcvMsg := string(bRcvMsg)
	th.Assert(t, rcvMsg == message, "expected the recieved message to be the same as the sent message.")

	// test context ends
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = SendToQueue(ctx, message, &mq, &ch, queueName)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected persistProducts to error out due to context ending")
}
