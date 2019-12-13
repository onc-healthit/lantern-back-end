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
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/onc-healthit/lantern-back-end/lanternmq"
)

type Job struct {
	Context      context.Context
	FHIRURL      *url.URL
	Client       *http.Client
	MessageQueue *lanternmq.MessageQueue
	Channel      *lanternmq.ChannelID
	QueueName    string
}

//var jobs = make(chan Job, 10)
//var results = make(chan Result, 10)

func worker(jobs chan Job, kill chan bool, wg *sync.WaitGroup) {
	for {
		select {
		case job := <-jobs:
			ctx, cancel := context.WithDeadline(job.Context, time.Now().Add(30*time.Second))
			defer cancel()
			GetAndSendCapabilityStatement(ctx, job.FHIRURL, job.Client, job.MessageQueue, job.Channel, job.QueueName)
			cancel()
		case <-kill:
			wg.Done()
			return
		}
	}
}

func CreateWorkerPool(noOfWorkers int, jobs chan Job, kill chan bool) *sync.WaitGroup {
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(jobs, kill, &wg)
	}
	return &wg
}
