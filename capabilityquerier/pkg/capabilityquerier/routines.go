package capabilityquerier

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
