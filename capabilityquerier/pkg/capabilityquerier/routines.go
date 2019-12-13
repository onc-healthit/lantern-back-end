package capabilityquerier

/*
want a pool of workers
want a channel that can hold as many jobs as there are workers (?)
  or perhaps just a normal channel is fine
want the pool of workers to stay active just until the jobs are done...
which means need a way to signal that jobs are done...
can't do this with just a single message on the queue because only one receiver will get it
a topic is async to the queue, so can't really do with a topic either...
we could say, when the queue is empty, quit. but what if it takes some time to put messages on the queue - more time than it takes to take them off
could send as many done messages as there are workers... and each worker stops receiving after a done message is received.
that seems reasonable...

future work:
receive messages from queue and have way of saying "this is a message" vs "this is done".
when start receiving messages, spin up the count of go routines.
shut everything down when receive "done" message using a kill message

now work:
on a given interval, spin up the count of go routines
send all urls over job channel
send kill signal when list of urls is gone
*/

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
)

type Job struct {
	Context      context.Context
	Duration     time.Duration
	FHIRURL      *url.URL
	Client       *http.Client
	MessageQueue *lanternmq.MessageQueue
	Channel      *lanternmq.ChannelID
	QueueName    string
}

type QueueWorkers struct {
	jobs       chan *Job
	kill       chan bool
	numWorkers int
	waitGroup  *sync.WaitGroup
	ctx        context.Context
}

func NewQueueWorkers() *QueueWorkers {
	qw := QueueWorkers{
		jobs:       make(chan *Job),
		kill:       make(chan bool),
		numWorkers: 0,
	}
	return &qw
}

func (qw *QueueWorkers) Start(ctx context.Context, numWorkers int) error {
	if qw.numWorkers > 0 {
		return errors.New("workers have already started")
	}
	var wg sync.WaitGroup
	qw.waitGroup = &wg
	qw.numWorkers = numWorkers
	qw.ctx = ctx
	for i := 0; i < qw.numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, qw.jobs, qw.kill, qw.waitGroup)
	}
	return nil
}

func (qw *QueueWorkers) Add(job *Job) error { // this checks if the context has completed before we start up the process
	select {
	case <-qw.ctx.Done():
		return qw.ctx.Err()
	default:
		// ok
	}

	qw.jobs <- job
	return nil
}

func (qw *QueueWorkers) Stop() error {
	select {
	case <-qw.ctx.Done():
		// wait for all the canceled workers to stop
		qw.waitGroup.Wait()
		qw.numWorkers = 0
		return nil
	default:
		// ok
	}

	if qw.numWorkers == 0 {
		return errors.New("no workers are currently running")
	}
	for i := 0; i < qw.numWorkers; i++ {
		qw.kill <- true
	}

	qw.waitGroup.Wait()

	qw.numWorkers = 0

	return nil
}

func worker(ctx context.Context, jobs chan *Job, kill chan bool, wg *sync.WaitGroup) {
	for {
		select {
		case job := <-jobs:
			jobCtx, cancel := context.WithDeadline(job.Context, time.Now().Add(job.Duration))
			GetAndSendCapabilityStatement(jobCtx, job.FHIRURL, job.Client, job.MessageQueue, job.Channel, job.QueueName)
			cancel()
		case <-ctx.Done():
			wg.Done()
			return
		case <-kill:
			wg.Done()
			return
		}
	}
}
