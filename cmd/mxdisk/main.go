package main

import (
	"fmt"
	"time"
	"github.com/maxxant/mxdisk"
)

func main() {
	printMntRemovableDisks := func(mp map[string]mxdisk.DiskInfo) {
		fmt.Printf("blk: %+v\n", mp)
	}

	timer := make(chan struct{})
	go func() {
		for {
			time.Sleep(time.Second * 10)
			timer <- struct{}{}
		}
	}()

	ch := mxdisk.WatchMounts(timer)

	for {
		select {
		case d, ok := <-ch:
			if !ok {
				return
			}
			printMntRemovableDisks(d)
		}
	}
}
