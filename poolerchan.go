package poolerchan

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var ErrPoolNotStarted = errors.New("poolerchan not started")

type Status uint8

const (
	Stopped Status = iota + 1
	Running
	Started
)

const (
	defaultNumberOfJobs    int = 8
	defaultNumberOfWorkers int = 4
)

type result struct {
	err error
}
type TaskQueued struct {
	t Task
}
type Task func(ctx context.Context) error

type Poolchan struct {
	status  Status
	options *Config

	jobQueue chan TaskQueued
	mtx      sync.Locker
}

type execute struct {
	*Poolchan
}

type Executer interface {
	Execute(context.Context) error
}

func NewPoolchan(opts ...ConfigOption) *Poolchan {
	pool := defaultConfigPoolchan()

	for _, option := range opts {
		option(pool)
	}

	return &Poolchan{
		status:   Stopped,
		options:  pool,
		mtx:      &sync.Mutex{},
		jobQueue: make(chan TaskQueued, pool.numberOfJobs),
	}
}

func (p *Poolchan) Queue(task Task) *Poolchan {
	if len(p.jobQueue) >= cap(p.jobQueue) {
		p.options.logger.Warn("job queue full")
		return p
	}
	p.jobQueue <- TaskQueued{
		t: task,
	}
	return p
}

func (p *Poolchan) Build() Executer {
	if p.options.numberOfWorkers > len(p.jobQueue) {
		p.options.logger.Warn("number of workers are more than number of tasks")
		p.options.numberOfWorkers = (len(p.jobQueue) % 2) + 1
	}
	d := *p
	d.status = Started
	return &execute{&d}
}

func (p *execute) Execute(ctx context.Context) error {
	p.mtx.Lock()
	if p.status != Started {
		p.mtx.Unlock()
		return ErrPoolNotStarted
	}
	p.status = Running
	p.mtx.Unlock()

	results := make(chan result, p.options.numberOfJobs)
	wg := &sync.WaitGroup{}

	for i := 0; i < p.options.numberOfWorkers; i++ {
		wg.Add(1)
		go p.executeWorker(ctx, results, wg)
	}

	close(p.jobQueue)

	var allErrors error
	go func() {
		wg.Wait()
		close(results)
	}()
	for i := 0; i < p.options.numberOfWorkers; i++ {
		res := <-results
		if res.err != nil {
			allErrors = errors.Join(allErrors, res.err)
		}
	}

	return allErrors
}

func (p *execute) executeWorker(ctx context.Context, res chan<- result, wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range p.jobQueue {
		p.options.logger.Debug("executing task")
		select {
		case <-ctx.Done():
			err := context.Cause(ctx)
			if errors.Is(err, context.Canceled) {
				// canceled
				fmt.Println("canceled")
			} else if errors.Is(err, context.DeadlineExceeded) {
				// deadline
				fmt.Println("deadline")
			} else {
				// another
				fmt.Println(err)
			}
			res <- result{err: err}
		default:
			res <- result{err: job.t(ctx)}
		}

	}
}
