package poolerchan

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"
)

type Config struct {
	numberOfJobs    int
	numberOfWorkers int

	context context.Context

	logger *log.Logger
}

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

type executePool struct {
	*Poolchan
}

type ExecuterPool interface {
	Execute(context.Context) error
}

func defaultConfigPoolchan() *Config {
	return &Config{
		numberOfJobs:    defaultNumberOfJobs,
		numberOfWorkers: defaultNumberOfWorkers,
		logger:          slog.NewLogLogger(slog.NewJSONHandler(os.Stderr, nil), slog.LevelInfo),
		context:         context.Background(),
	}
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
		p.options.logger.Println("job queue full")
		return p
	}
	p.jobQueue <- TaskQueued{
		t: task,
	}
	return p
}

func (p *Poolchan) Build() ExecuterPool {
	d := *p
	d.status = Started
	return &executePool{&d}
}

func (p *executePool) Execute(ctx context.Context) error {
	p.mtx.Lock()
	if p.status != Started {
		p.mtx.Unlock()
		return fmt.Errorf("pool not started")
	}
	p.status = Running
	p.mtx.Unlock()

	results := make(chan result, p.options.numberOfJobs)

	for i := 0; i < p.options.numberOfWorkers; i++ {
		go p.executeWorker(ctx, results)
	}

	close(p.jobQueue)
	defer close(results)

	var allErrors error
	for i := 0; i < p.options.numberOfJobs; i++ {
		res := <-results
		if res.err != nil {
			allErrors = errors.Join(allErrors, res.err)
		}
	}

	return allErrors
}

func (p *executePool) executeWorker(ctx context.Context, res chan<- result) {
	for job := range p.jobQueue {
		res <- result{err: job.t(ctx)}
	}
}
