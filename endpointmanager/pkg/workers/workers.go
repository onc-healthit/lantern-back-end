package workers

import (
	"context"
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// Job contains all of the information for a worker to execute the job.
// A job contains a context and a duration. The job handler is provided a new context
// for the job based off or the job's provided context and the given duration.
type Job struct {
	Context     context.Context
	Duration    time.Duration
	Handler     func(context.Context, *map[string]interface{}) error
	HandlerArgs *map[string]interface{}
}

// Workers handles the provided number of workers and allows jobs to be sent to the
// workers and distributes those jobs to the workers.
type Workers struct {
	jobs       chan *Job
	kill       chan bool
	numWorkers int
	waitGroup  *sync.WaitGroup
	ctx        context.Context
}

// NewWorkers initializes a QueueWorkers structure.
func NewWorkers() *Workers {
	w := Workers{
		jobs:       make(chan *Job),
		kill:       make(chan bool),
		numWorkers: 0,
	}
	return &w
}

// Start creates the number of workers provided. It also runs using a context. If the context ends,
// a signal is sent to each worker to stop working after they have completed their latest job.
// Start throws an error if Workers have already been started and has not been stopped.
func (w *Workers) Start(ctx context.Context, numWorkers int, errs chan error) error {
	if w.numWorkers > 0 {
		return errors.New("workers have already started")
	}
	var wg sync.WaitGroup
	w.waitGroup = &wg
	w.numWorkers = numWorkers
	w.ctx = ctx
	for i := 0; i < w.numWorkers; i++ {
		wg.Add(1)
		go worker(ctx, w.jobs, w.kill, w.waitGroup, errs)
	}
	return nil
}

// Add takes a Job as an argument and sends that job to the workers to be executed when a
// worker is available.
func (w *Workers) Add(job *Job) error { // this checks if the context has completed before we start up the process
	if w.numWorkers == 0 {
		return errors.New("no workers are currently running")
	}
	select {
	case <-w.ctx.Done():
		return w.ctx.Err()
	default:
		// ok
	}

	w.jobs <- job
	return nil
}

// Stop sends a stop signal to all of the workers to stop accepting jobs and to close.
// Stop throws an error if QueueWorkers has already been stopped and has not been restarted.
func (w *Workers) Stop() error {
	if w.numWorkers == 0 {
		return errors.New("no workers are currently running")
	}

	select {
	case <-w.ctx.Done():
		// wait for all the canceled workers to stop
		w.waitGroup.Wait()
		w.numWorkers = 0
		return nil
	default:
		// ok
	}

	for i := 0; i < w.numWorkers; i++ {
		w.kill <- true
	}

	w.waitGroup.Wait()

	w.numWorkers = 0
	return nil
}

func jobHandler(job *Job) error {
	jobCtx, cancel := context.WithDeadline(job.Context, time.Now().Add(job.Duration))
	defer cancel()

	err := job.Handler(jobCtx, job.HandlerArgs)
	return err
}

func worker(ctx context.Context, jobs chan *Job, kill chan bool, wg *sync.WaitGroup, errs chan<- error) {
	for {
		select {
		case job := <-jobs:
			err := jobHandler(job)
			log.Warnf("JOBS DONE")
			if err != nil {
				errs <- err
			}
		case <-ctx.Done():
			log.Warnf("CONTEXT DONE %+v", job.HandlerArgs)
			wg.Done()
			return
		case <-kill:
			log.Warnf("KILLED %+v", job.HandlerArgs)
			wg.Done()
			return
		}
	}
}
