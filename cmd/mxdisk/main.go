package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/maxxant/mxdisk"
)

func main() {
	done := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt) // Ctrl+C
	go func() {
		<-c
		//fmt.Println("\nReceived an interrupt, stopping services...")
		done <- struct{}{}
	}()

	ch := mxdisk.Watch(done, mxdisk.NewConfig(), true)

	for {
		select {
		case d, ok := <-ch:
			if !ok {
				return
			}
			fmt.Print(d)
		}
	}
}
