package mxdisk

import (
	"fmt"
	"github.com/maxxant/udev"
	"strings"
)

func WatchUdev() {
	monitor, err := udev.NewMonitor()
	if nil != err {
		fmt.Println(err)
		return
	}

	defer monitor.Close()
	events := make(chan *udev.UEvent)
	monitor.Monitor(events)
	for {
		event := <-events

		if devt, ok := event.Env["DEVTYPE"]; ok {
			if devt == "disk" || devt == "partition" {
				//fmt.Println(event.String())
				name := strings.Split(event.Devpath, "/")
				name = name[len(name)-1:]
				fmt.Println(event.Action, name, devt)
			}
		}
	}
}
