package capabilityquerier

import (
	"context"
	"fmt"
	"net/url"
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

	jctx := context.Background()
	duration := 30 * time.Second

	fhirURL := &url.URL{}
	fhirURL, err = fhirURL.Parse(sampleURL)
	th.Assert(t, err == nil, err)

	tc, err := testClientWithContentType(fhir2LessJSONMIMEType)
	th.Assert(t, err == nil, err)
	defer tc.Close()

	job := Job{
		Context:      jctx,
		Duration:     duration,
		FHIRURL:      fhirURL,
		Client:       &(tc.Client),
		MessageQueue: &mq,
		Channel:      &ch,
		QueueName:    queueName,
	}

	// basic test

	// start workers
	ctx := context.Background()
	numWorkers := 3

	qw := NewQueueWorkers()
	err = qw.Start(ctx, numWorkers)
	th.Assert(t, err == nil, err)
	th.Assert(t, qw.waitGroup != nil, "expected a wait group to be initiated")
	th.Assert(t, qw.numWorkers == numWorkers, fmt.Sprintf("should have %d workers; have %d", numWorkers, qw.numWorkers))
	th.Assert(t, qw.ctx == ctx, "queue workers context should be the same as the passed in context")

	for i := 0; i < numWorkers*2; i++ {
		err = qw.Add(&job)
		th.Assert(t, err == nil, err)
	}

	// expect to have not al items on queue because added more jobs than there are workers.
	numOnQueue := len(mq.(*mock.BasicMockMessageQueue).Queue)
	th.Assert(t, numOnQueue < numWorkers*2, fmt.Sprintf("expected less than %d items to be on the queue. had %d.", 2*numWorkers, numOnQueue))

	err = qw.Stop()
	th.Assert(t, err == nil, err)
	th.Assert(t, qw.numWorkers == 0, "after stopping, there should be no workers")

	// expect all items to be on queue after stopped
	numOnQueue = len(mq.(*mock.BasicMockMessageQueue).Queue)
	th.Assert(t, numOnQueue == numWorkers*2, fmt.Sprintf("expected %d items to be on the queue. had %d.", 2*numWorkers, numOnQueue))

	// stop after already stopped

	err = qw.Stop()
	th.Assert(t, err.Error() == "no workers are currently running", "expected error saying no workers are running.")

	// start after already started

	err = qw.Start(ctx, numWorkers)
	th.Assert(t, err == nil, err)
	err = qw.Start(ctx, numWorkers)
	th.Assert(t, err.Error() == "workers have already started", "expected error saying workers were already running.")
	err = qw.Stop()
	th.Assert(t, err == nil, err)

	// canceled context

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// expect it to start fine
	err = qw.Start(ctx, numWorkers)
	th.Assert(t, err == nil, err)

	// expect error when adding a new job
	err = qw.Add(&job)
	th.Assert(t, errors.Cause(err) == context.Canceled, "expected to error out due to context ending")
	// shouldn't have anything new on the queue
	th.Assert(t, numOnQueue == numWorkers*2, fmt.Sprintf("expected %d items to be on the queue. had %d.", 2*numWorkers, numOnQueue))

	// expect no issues with stopping
	err = qw.Stop()
	th.Assert(t, err == nil, err)
	th.Assert(t, qw.numWorkers == 0, "after stopping, there should be no workers")
}
