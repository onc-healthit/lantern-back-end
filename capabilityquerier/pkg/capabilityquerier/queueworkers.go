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

// Job contains all of the information for a queue worker to execute the job.
// A job contains a context and a duration. The job handler is provided a new context
// for the job based off or the job's provided context and the given duration.
// TODO: if queue workers make sense in multiple areas of the code, this should
// be abstracted to 'interface{}' or similar so arbitrary jobs can be sent.
type Job struct {
	Context      context.Context
	Duration     time.Duration
	FHIRURL      *url.URL
	Client       *http.Client
	MessageQueue *lanternmq.MessageQueue
	Channel      *lanternmq.ChannelID
	QueueName    string
}

// QueueWorkers handles the provided number of queue workers and allows jobs to be sent to the queue
// workers and distributes those jobs to the queue workers.
type QueueWorkers struct {
	jobs       chan *Job
	kill       chan bool
	numWorkers int
	waitGroup  *sync.WaitGroup
	ctx        context.Context
}

// NewQueueWorkers initializes a QueueWorkers structure.
func NewQueueWorkers() *QueueWorkers {
	qw := QueueWorkers{
		jobs:       make(chan *Job),
		kill:       make(chan bool),
		numWorkers: 0,
	}
	return &qw
}

// Start creates the number of workers provided. It also runs using a context. If the context ends,
// a signal is sent to each worker to stop working after they have completed their latest job.
// Start throws an error if QueueWorkers has already been started and has not been stopped.
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

// Add takes a Job as an argument and sends that job to the workers to be executed when a
// worker is available.
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

// Stop sends a stop signal to all of the workers to stop accepting jobs and to close.
// Stop throws an error if QueueWorkers has already been stopped and has not been restarted.
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
