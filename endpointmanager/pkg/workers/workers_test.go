package workers

import (
	"context"
	"fmt"
	"testing"
	"time"

	th "github.com/onc-healthit/lantern-back-end/endpointmanager/pkg/testhelper"
	"github.com/onc-healthit/lantern-back-end/lanternmq"
	"github.com/onc-healthit/lantern-back-end/lanternmq/mock"
	"github.com/pkg/errors"
)

func Test_StartAddAndStop(t *testing.T) {
	// setup

	var err error
	var ch lanternmq.ChannelID

	mq := mock.NewBasicMockMessageQueue()
	ch = 1
	queueName := "queue name"
	errs := make(chan error)

	jctx := context.Background()
	duration := 30 * time.Second
	args := make(map[string]interface{})
	args["mq"] = mq
	args["ch"] = ch
	args["queueName"] = queueName

	job := Job{
		Context:     jctx,
		Duration:    duration,
		Handler:     testfn,
		HandlerArgs: &args,
	}

	// basic test

	// start workers
	ctx := context.Background()
	numWorkers := 3

	work := NewWorkers()
	err = work.Start(ctx, numWorkers, errs)
	th.Assert(t, err == nil, err)
	th.Assert(t, work.waitGroup != nil, "expected a wait group to be initiated")
	th.Assert(t, work.numWorkers == numWorkers, fmt.Sprintf("should have %d workers; have %d", numWorkers, work.numWorkers))
	th.Assert(t, work.ctx == ctx, "queue workers context should be the same as the passed in context")

	for i := 0; i < numWorkers*2; i++ {
		err = work.Add(&job)
		th.Assert(t, err == nil, err)
	}

	// expect to have less than or equal to all items on queue because added more jobs than there are workers.
	numOnQueue := len(mq.(*mock.BasicMockMessageQueue).Queue)
	th.Assert(t, numOnQueue <= numWorkers*2, fmt.Sprintf("expected less than or equal to %d items to be on the queue. had %d.", 2*numWorkers, numOnQueue))

	err = work.Stop()
	th.Assert(t, err == nil, err)
	th.Assert(t, work.numWorkers == 0, "after stopping, there should be no workers")

	// expect all items to be on queue after stopped
	numOnQueue = len(mq.(*mock.BasicMockMessageQueue).Queue)
	th.Assert(t, numOnQueue == numWorkers*2, fmt.Sprintf("expected %d items to be on the queue. had %d.", 2*numWorkers, numOnQueue))

	// stop after already stopped

	err = work.Stop()
	th.Assert(t, err.Error() == "no workers are currently running", "expected error saying no workers are running.")

	// start after already started

	err = work.Start(ctx, numWorkers, errs)
	th.Assert(t, err == nil, err)
	err = work.Start(ctx, numWorkers, errs)
	th.Assert(t, err.Error() == "workers have already started", "expected error saying workers were already running.")
	err = work.Stop()
	th.Assert(t, err == nil, err)

	// canceled context

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// expect it to start fine
	err = work.Start(ctx, numWorkers, errs)
	th.Assert(t, err == nil, err)

	// expect error when adding a new job
	err = work.Add(&job)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected to error out due to context ending")
	// shouldn't have anything new on the queue
	th.Assert(t, numOnQueue == numWorkers*2, fmt.Sprintf("expected %d items to be on the queue. had %d.", 2*numWorkers, numOnQueue))

	// expect no issues with stopping
	err = work.Stop()
	th.Assert(t, err == nil, err)
	th.Assert(t, work.numWorkers == 0, "after stopping, there should be no workers")
}

// testfn is an example handler function for the Job to run that just sends a test string over a queue
func testfn(ctx context.Context, args *map[string]interface{}) error {
	mq, ok := (*args)["mq"].(lanternmq.MessageQueue)
	if !ok {
		return fmt.Errorf("unable to cast mq to MessageQueue from arguments")
	}
	ch, ok := (*args)["ch"].(lanternmq.ChannelID)
	if !ok {
		return fmt.Errorf("unable to cast ch to ChannelID from arguments")
	}
	queueName, ok := (*args)["queueName"].(string)
	if !ok {
		return fmt.Errorf("unable to cast queueName to string from arguments")
	}
	err := mq.PublishToQueue(ch, queueName, "test String")
	if err != nil {
		return err
	}
	return nil
}
