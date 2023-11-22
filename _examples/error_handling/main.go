package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ervitis/poolerchan"
)

func main() {
	const sec = 15
	ctx, cancel := context.WithTimeout(context.Background(), sec*time.Second)
	defer func() {
		cancel()
	}()
	myQueueBuilder := poolerchan.NewPoolchan(poolerchan.WithContext(ctx))

	queue := myQueueBuilder.
		Queue(func(ctx context.Context) error {
			fmt.Println("hello world")
			return nil
		}).
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my second task")
			return nil
		}).
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my third task")
			return nil
		}).
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my forth task")
			return errors.New("in my forth task i fail")
		}).
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my fifth task")
			return nil
		}).
		Queue(func(ctx context.Context) error {
			fmt.Println("this is my sixth task")
			return nil
		}).
		Build()
	if err := queue.Execute(ctx); err != nil {
		log.Println(err)
	}
	fmt.Println("end!")
}
