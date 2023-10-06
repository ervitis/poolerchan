package main

import (
	"context"
	"fmt"
	"github.com/ervitis/poolerchan"
	"log"
)

func main() {
	myQueueBuilder := poolerchan.NewPoolchan()

	queue := myQueueBuilder.
		Queue(func(ctx context.Context) error {
			fmt.Println("hello world")
			return nil
		}).
		Build()
	log.Println(queue.Execute())
}
