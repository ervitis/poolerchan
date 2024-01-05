package poolerchan

import (
	"bytes"
	"context"
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
