package poolerchan

import (
	"context"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
	"os"

	"golang.org/x/sync/semaphore"
)

//go:generate go run github.com/ervitis/foggo@latest afop --struct Pool --no-instance
type Config struct {
	numberOfJobs    int
	numberOfWorkers int

	logger *log.Logger
	ctx    context.Context
}

type Status uint8

const (
	Stopped Status = iota + 1
	Running
	Started
)

type Task func(ctx context.Context) error

type Poolchan struct {
	status  Status
	options *Config

	eg       *errgroup.Group
	jobQueue chan Task
	sem      *semaphore.Weighted
}

type executePool struct {
	*Poolchan
}

type ExecuterPool interface {
	Execute() error
}

func defaultConfigPoolchan() *Config {
	return &Config{
		numberOfJobs:    5,
		numberOfWorkers: 3,
		logger:          slog.NewLogLogger(slog.NewJSONHandler(os.Stderr, nil), slog.LevelInfo),
		ctx:             context.Background(),
	}
}

func NewPoolchan(opts ...PoolOption) *Poolchan {
	pool := defaultConfigPoolchan()

	for _, option := range opts {
		option.apply(pool)
	}

	g, ctx := errgroup.WithContext(pool.ctx)
	pool.ctx = ctx

	return &Poolchan{
		eg:       g,
		status:   Stopped,
		options:  pool,
		jobQueue: make(chan Task),
		sem:      semaphore.NewWeighted(int64(pool.numberOfWorkers)),
	}
}

func (p *Poolchan) Queue(task Task) *Poolchan {
	p.jobQueue <- task
	return p
}

func (p *Poolchan) Build() ExecuterPool {
	d := *p
	d.status = Started
	return &executePool{&d}
}

func (p *executePool) Execute() error {
	// acquire semaphore
	// check status

	// check jobs if they are available in jobqueue
	// get jobs and put them in the jobQueueRunning and execute them in parallel with errgroup

	// wait them to finish
	// when one job finish, get jobs from jobqueue and do it again

	return nil
}
