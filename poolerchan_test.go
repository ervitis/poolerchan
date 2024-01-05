package poolerchan

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

func TestPoolchan_Queue(t *testing.T) {
	type args struct {
		tasks []Task
	}
	tests := []struct {
		name        string
		args        args
		cfg         []ConfigOption
		checkResult func(p *Poolchan, args args) error
	}{
		{
			name: "success",
			args: args{
				tasks: []Task{
					func(ctx context.Context) error {
						return nil
					},
					func(ctx context.Context) error {
						return nil
					},
					func(ctx context.Context) error {
						return nil
					},
				},
			},
			checkResult: func(p *Poolchan, arg args) error {
				var err error
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("recovered from panic: %v", r)
					}
				}()
				for _, v := range arg.tasks {
					p.Queue(v)
				}
				return err
			},
		},
		{
			name: "queue full",
			args: args{
				tasks: []Task{
					func(ctx context.Context) error {
						return nil
					},
					func(ctx context.Context) error {
						return nil
					},
					func(ctx context.Context) error {
						return nil
					},
					func(ctx context.Context) error {
						return nil
					},
				},
			},
			cfg: []ConfigOption{
				WithNumberOfJobs(1),
			},
			checkResult: func(p *Poolchan, args args) error {
				buff := &bytes.Buffer{}
				p.options.logger = slog.New(slog.NewTextHandler(buff, nil))

				tasks := args.tasks
				p.Queue(tasks[0]).Queue(tasks[1]).Queue(tasks[2])

				msg := strings.Split(buff.String(), "\n")[0]

				if !strings.Contains(msg, "job queue full") {
					return fmt.Errorf("queue full but no message is printed in logger")
				}
				if len(p.jobQueue) != 1 {
					return fmt.Errorf("job queue len not correct")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPoolchan(tt.cfg...)
			if tt.checkResult == nil {
				t.Fail()
			}
			if err := tt.checkResult(p, tt.args); err != nil {
				t.Errorf("error ocurred checking result in test: %v", err)
			}
		})
	}
}

func TestPoolchan_Build(t *testing.T) {
	type args struct {
		tasks []Task
	}
	tests := []struct {
		name        string
		args        args
		cfg         []ConfigOption
		checkResult func(p *Poolchan, args args) error
	}{
		{
			name: "types correct",
			cfg:  []ConfigOption{WithNumberOfJobs(1)},
			checkResult: func(p *Poolchan, args args) error {
				pooler := p.Queue(func(ctx context.Context) error {
					return nil
				}).Build()
				if _, ok := pooler.(Executer); !ok {
					return fmt.Errorf("pooler not type Executer")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPoolchan(tt.cfg...)
			if tt.checkResult == nil {
				t.Fail()
			}
			if err := tt.checkResult(p, tt.args); err != nil {
				t.Errorf("error ocurred checking result in test: %v", err)
			}
		})
	}
}

type customErr struct {
	msg string
}

func (e *customErr) Error() string {
	return e.msg
}

func Test_execute_Execute(t *testing.T) {
	var buff = new(bytes.Buffer)

	globalLogger := slog.New(slog.NewTextHandler(buff, nil))

	cErr := &customErr{msg: "ups"}

	type args struct {
		ctx   context.Context
		tasks []Task
	}
	tests := []struct {
		name        string
		args        args
		cfg         []ConfigOption
		checkResult func(p *Poolchan, args args) error
	}{
		{
			name: "success",
			cfg:  []ConfigOption{WithNumberOfJobs(2)},
			args: args{
				ctx: context.TODO(),
				tasks: []Task{
					func(ctx context.Context) error {
						globalLogger.Info("task1")
						return nil
					},
					func(ctx context.Context) error {
						globalLogger.Info("task2")
						return nil
					},
				},
			},
			checkResult: func(p *Poolchan, args args) error {
				tasks := args.tasks
				pooler := p.Queue(tasks[0]).Queue(tasks[1]).Build()
				if err := pooler.Execute(args.ctx); err != nil {
					return fmt.Errorf("error executing pool worker: %w", err)
				}
				if buff.Len() == 0 {
					return nil
				}
				msg := strings.SplitN(buff.String(), "\n", 2)
				if len(msg) <= 1 {
					return fmt.Errorf("not printed the tasks")
				}

				c := 0
				for i := range msg {
					if strings.Contains(msg[i], "task1") {
						c++
					}
					if strings.Contains(msg[i], "task2") {
						c++
					}
				}
				if c != 2 {
					return fmt.Errorf("tasks not run")
				}
				return nil
			},
		},
		{
			name: "succes, but with errors",
			cfg:  []ConfigOption{WithNumberOfJobs(2)},
			args: args{
				ctx: context.TODO(),
				tasks: []Task{
					func(ctx context.Context) error {
						globalLogger.Info("task1")
						return new(customErr)
					},
					func(ctx context.Context) error {
						globalLogger.Info("task2")
						return nil
					},
				},
			},
			checkResult: func(p *Poolchan, args args) error {
				tasks := args.tasks
				pooler := p.Queue(tasks[0]).Queue(tasks[1]).Build()
				err := pooler.Execute(args.ctx)
				if err == nil {
					return fmt.Errorf("should be an error executing pool worker")
				}
				if buff.Len() == 0 {
					return nil
				}
				msg := strings.SplitN(buff.String(), "\n", 2)
				if len(msg) <= 1 {
					return fmt.Errorf("not printed the tasks")
				}

				c := 0
				for i := range msg {
					if strings.Contains(msg[i], "task1") {
						c++
					}
					if strings.Contains(msg[i], "task2") {
						c++
					}
				}
				if c != 2 {
					return fmt.Errorf("tasks not run")
				}

				if !errors.As(err, &cErr) {
					return fmt.Errorf("i cannot find the custom error")
				}

				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff.Reset()
			p := NewPoolchan(tt.cfg...)
			if tt.checkResult == nil {
				t.Fail()
			}
			if err := tt.checkResult(p, tt.args); err != nil {
				t.Errorf("error ocurred checking result in test: %v", err)
			}
		})
	}
}
