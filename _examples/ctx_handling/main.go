package main

import (
	"context"
	"fmt"
	"github.com/ervitis/poolerchan"
	"log"
	"time"
)

func main() {
	const sec = 1
	ctx, cancel := context.WithTimeout(context.Background(), sec*time.Second)
	defer func() {
		cancel()
	}()
	myQueueBuilder := poolerchan.NewPoolchan(
		poolerchan.WithNumberOfWorkers(6),
	)

	// TODO it fails with deadlock if number of workers > number of queued tasks
	// i have to control the results in the loop when retrieving the results
	queue := myQueueBuilder.
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my first task")
			time.Sleep(2 * time.Second)
			cancel()
			return nil
		}).
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my s task")
			time.Sleep(2 * time.Second)
			cancel()
			return nil
		}).
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my t task")
			time.Sleep(2 * time.Second)
			cancel()
			return nil
		}).
		Build()
	if err := queue.Execute(ctx); err != nil {
		log.Println(err)
	}
	fmt.Println("end!")
}
