package poolerchan

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
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

const defaultMinQueueLen int = 16

type Result struct {
	err error
}
type TaskQueued struct {
	t Task
}
type Task func(ctx context.Context) error

type Poolchan struct {
	status  Status
	options *Config

	eg       *errgroup.Group
	jobQueue chan TaskQueued
	sem      *semaphore.Weighted
	mtx      sync.Locker
	err      chan error
}

type executePool struct {
	*Poolchan
}

type ExecuterPool interface {
	Execute(context.Context) error
}

func defaultConfigPoolchan() *Config {
	return &Config{
		numberOfJobs:    5,
		numberOfWorkers: 3,
		logger:          slog.NewLogLogger(slog.NewJSONHandler(os.Stderr, nil), slog.LevelInfo),
		context:         context.Background(),
	}
}

func NewPoolchan(opts ...ConfigOption) *Poolchan {
	pool := defaultConfigPoolchan()

	for _, option := range opts {
		option(pool)
	}

	g, ctx := errgroup.WithContext(pool.context)
	pool.context = ctx

	return &Poolchan{
		eg:       g,
		status:   Stopped,
		options:  pool,
		mtx:      &sync.Mutex{},
		err:      make(chan error),
		jobQueue: make(chan TaskQueued, pool.numberOfJobs),
		sem:      semaphore.NewWeighted(int64(pool.numberOfWorkers)),
	}
}

func (p *Poolchan) Queue(task Task) *Poolchan {
	if len(p.jobQueue) >= cap(p.jobQueue) {
		log.Println("job queue full")
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
	p.status = Started
	p.mtx.Unlock()

	wg := &sync.WaitGroup{}
	for i := 0; i < p.options.numberOfWorkers; i++ {
		wg.Add(1)
		go func() {
			executeWorker(ctx, p.jobQueue, p.err, p.sem)
			wg.Done()
		}()
	}
	wg.Wait()

	defer func() {
		close(p.jobQueue)
		close(p.err)
	}()

	for err := range p.err {
		log.Println(err)
	}

	return nil
}

func executeWorker(ctx context.Context, taskAcquired <-chan TaskQueued, errCh chan error, sem *semaphore.Weighted) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-taskAcquired:
			if !ok {
				return
			}
			if err := sem.Acquire(ctx, 1); err != nil {
				panic(err)
			}
			if err := task.t(ctx); err != nil {
				go func() {
					errCh <- err
				}()
			}
			sem.Release(1)
		}
	}
}
