package main

import (
	"fmt"
	"github.com/maxxant/mxdisk"
	"os"
	"os/signal"
)

func main() {
	done := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		//fmt.Println("\nReceived an interrupt, stopping services...")
		done <- struct{}{}
	}()

	ch := mxdisk.WatchMounts(done, mxdisk.NewConfig(), true)

	go mxdisk.WatchUdev()

	for {
		select {
		case d, ok := <-ch:
			if !ok {
				return
			}
			fmt.Println(d)
		}
	}
}
