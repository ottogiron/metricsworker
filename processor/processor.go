package processor

import (
	"context"
	"fmt"
	"sync"
	"time"

	fworkerprocessor "github.com/ferrariframework/ferrariworker/processor"
	"github.com/ottogiron/metricsworker/worker"
)

//Processor represents a tasks processor. It passes tasks to workers to execute business logic
type Processor interface {
	Register(id string, worker worker.Worker)
	Start() error
}

type taskResult struct {
	err      error
	workerID string
}

var _ Processor = (*processor)(nil)

type processor struct {
	adapter fworkerprocessor.Adapter
	//Number of registered workers running concurrently
	concurrency int
	//Time the processor will wait until new tasks are available
	waitTimeout    time.Duration
	workerRegistry map[string]worker.Worker
}

//NewProcessor returns a new instance of a processor
func NewProcessor(adapter fworkerprocessor.Adapter, options ...Option) Processor {
	//Initialize and set defaults
	p := &processor{
		concurrency: 1,
		waitTimeout: time.Millisecond * 500,
		adapter:     adapter,
	}

	//Apply user defined options
	for _, option := range options {
		option(p)
	}
	return p
}

//Start starts the task processor
func (p *processor) Start() error {
	//open the connection
	err := p.adapter.Open()
	if err != nil {
		return fmt.Errorf("Failed to open the processor Adapter connection %s", err)
	}
	defer p.adapter.Close()
	wg := sync.WaitGroup{}
	//Wait for the timeout once then call done to exit the processing
	wg.Add(p.concurrency)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgs, err := p.adapter.Messages(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get messages from adapter %s", err)

	}
	for i := 0; i < p.concurrency; i++ {
		go func() {
			for {
				select {
				case m, ok := <-msgs:
					if ok {
						ids := make([]string, 0, len(p.workerRegistry))
						for id := range p.workerRegistry {
							ids = append(ids, id)
						}
						out := p.process(m.OriginalMessage, ids...)

						for taskResult := range out {
							if taskResult.err != nil {
								fmt.Printf("Error Failed to execute task for worker id: %s in first attempt retrying", taskResult.workerID)
							}
						}

					} else {
						wg.Done()
						return
					}
				case <-time.After(p.waitTimeout * time.Millisecond):

					wg.Done()
					return
				}
			}
		}()
	}
	wg.Wait()
	return nil
}

func (p *processor) handleFailedTask(taskResult *taskResult) {
	fmt.Print("Do something with this failed task", taskResult)
}

//Process will process a task in all the available workers asynchronously
func (p *processor) process(task interface{}, workersIDS ...string) <-chan taskResult {
	out := make(chan taskResult)
	var wg sync.WaitGroup

	for _, id := range workersIDS {
		w := p.workerRegistry[id]
		go func(w worker.Worker, workerID string) {
			err := w.Execute(task)
			out <- taskResult{workerID: workerID, err: err}
		}(w, id)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

//Register register a new worker to execute a task
func (p *processor) Register(id string, worker worker.Worker) {
	p.workerRegistry[id] = worker
}
