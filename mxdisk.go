package mxdisk

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/maxxant/udev" // vendor fork from: github.com/deniswernert/udev
)

// Watch return chan with removable storage info
// onlyUUID for mounted devs with UUID only for filtering /dev/loop, etc
func Watch(done chan struct{}, config *Config, onlyUUID bool) chan DisksSummaryMap {
	udevDisks := newUdevMapInfo()
	fmt.Println("udevDisks:")
	fmt.Println(udevDisks)

	mounts := mapMntFile("/proc/mounts", udevDisks)
	fmt.Println("mounts:")
	fmt.Println(mounts)

	mblk := fetchSysBlock("/sys/class/block")
	fmt.Println("sysblock:")
	fmt.Println(mblk)

	fstab := mapMntFile("/etc/fstab", udevDisks)
	fstabandslaves := mblk.exposeDevsSlaves(fstab.devPaths())
	ft := newFstabMap(fstab, fstabandslaves)
	fmt.Println("fstabEx:")
	fmt.Println(ft)

	resMap := newDisksSummaryMap()
	resMap.rebuild(mblk)
	resMap.mergeFstabMap(ft)
	resMap.mergeMntMap(mounts)
	resMap.mergeUdevMap(udevDisks)

	events := make(chan *udev.UEvent)
	monitor, err := udev.NewMonitor()
	if nil != err {
		fmt.Println(err)
		// TODO additionals steps ?
	} else {
		monitor.Monitor(events)
	}

	timer := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(config.MonitoringFstabSec))
			timer <- true
		}
	}()

	fmt.Println("disks:")

	rch := make(chan DisksSummaryMap)
	go func() {
		rch <- resMap
		for {

			// make a copy for compare later
			oldMap := make(DisksSummaryMap, len(resMap))
			for k, v := range resMap {
				oldMap[k] = v
			}

			select {
			// mnt monitoring
			case <-time.After(time.Second * time.Duration(config.MonitoringProcmountSec)):
				udevDisks = newUdevMapInfo()
				// rescan ft
				resMap.mergeUdevMap(udevDisks)
				mounts = mapMntFile("/proc/mounts", udevDisks)
				resMap.mergeMntMap(mounts)
				if !reflect.DeepEqual(resMap, oldMap) {
					rch <- resMap
				}

			// udev monitoring
			case event, ok := <-events:
				if ok {
					if devt, ok := event.Env["DEVTYPE"]; ok {
						if devt == "disk" || devt == "partition" {
							//fmt.Println(event.String())
							name := strings.Split(event.Devpath, "/")
							name = name[len(name)-1:]
							fmt.Println(event.Action, name, devt)

							mblk = fetchSysBlock("/sys/class/block")
							resMap.rebuild(mblk)
							if !reflect.DeepEqual(resMap, oldMap) {
								rch <- resMap
							}
						}
					}
				}

			// fstab monitoring (optional disabled)
			case <-timer:
				// for next scan mnt tick
				fstab = mapMntFile("/etc/fstab", udevDisks)
				fstabandslaves = mblk.exposeDevsSlaves(fstab.devPaths())
				ft = newFstabMap(fstab, fstabandslaves)
				resMap.mergeFstabMap(ft)

			case <-done:
				close(rch)
				monitor.Close()
				return
			}
		}
	}()
	return rch
}
