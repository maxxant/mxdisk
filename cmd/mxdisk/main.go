package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/maxxant/mxdisk"
)

//var getDataCh chan string
var forceupCh chan struct{}
var ch chan mxdisk.DisksSummaryMap
var done chan struct{}

func init() {
	//getDataCh = make(chan string)
	forceupCh = make(chan struct{})
	done = make(chan struct{})
}

// RunHal main process
func RunHal() {
	//done := make(chan struct{})

	//c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt) // Ctrl+C
	// go func() {
	// 	<-c
	// 	//fmt.Println("\nReceived an interrupt, stopping services...")
	// 	done <- struct{}{}
	// }()

	ch = mxdisk.Watch(done, mxdisk.NewConfig(), true, forceupCh)
}

// GetLastJSONData get last json data
func GetLastJSONData() string {
	go func() {
		select {
		case forceupCh <- struct{}{}:
		case <-time.After(time.Second):
			// deadline
			break
		}
	}()

	select {
	case d, ok := <-ch:
		if ok {
			r := mxdisk.NewDiskMap(d)
			r.FilterFstab()
			r.FilterVirtual()
			r.FillFsTypeIfEmpty()
			r.FillDevIDs()
			ba, err := json.Marshal(r)
			if err == nil {
				return string(ba)
			}
		}

	case <-time.After(time.Second * 2):
		// emergency deadline
		break
	}

	fmt.Println("GetLastJSONData empty")
	return string("")
}

var initHalStorageInit bool

func init() {
	initHalStorageInit = false
}

//export goHalStorageInit
func goHalStorageInit() {
	//C.reMyHalStorage()
	if !initHalStorageInit {
		initHalStorageInit = true
		RunHal()
	}
}

//export goHalStorageGetJSON
func goHalStorageGetJSON() string {
	return GetLastJSONData()
}

func main() {
	goHalStorageInit()

	for {
		time.Sleep(time.Second)
		fmt.Print(goHalStorageGetJSON())
	}
	// done := make(chan struct{})
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt) // Ctrl+C
	// go func() {
	// 	<-c
	// 	//fmt.Println("\nReceived an interrupt, stopping services...")
	// 	done <- struct{}{}
	// }()

	// ch := mxdisk.Watch(done, mxdisk.NewConfig(), true, nil)

	// for {
	// 	select {
	// 	case d, ok := <-ch:
	// 		if !ok {
	// 			return
	// 		}
	// 		// fmt.Println("event")
	// 		// fmt.Print(d)
	// 		r := mxdisk.NewDiskMap(d)
	// 		r.FilterFstab()
	// 		r.FilterVirtual()
	// 		r.FillFsTypeIfEmpty()
	// 		r.FillDevIDs()
	// 		b, _ := json.Marshal(r)
	// 		fmt.Print(string(b) + "\n")
	// 	}
	// }
}
