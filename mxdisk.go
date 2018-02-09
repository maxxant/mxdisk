package mxdisk

import (
	"fmt"
	"time"
	//"github.com/maxxant/udev" // vendor fork from: github.com/deniswernert/udev
)

// Watch return chan with removable storage info
// onlyUUID for mounted devs with UUID only for filtering /dev/loop, etc
func Watch(done chan struct{}, config *Config, onlyUUID bool, forceUpdate <-chan struct{}) chan DisksSummaryMap {
	udevDisks := newUdevMapInfo()
	mounts := mapMntFile("/proc/mounts", udevDisks)
	mblk := fetchSysBlock("/sys/class/block")
	fstab := mapMntFile("/etc/fstab", udevDisks)
	fstabandslaves := mblk.exposeDevsSlaves(fstab.devPaths())
	ft := newFstabMap(fstab, fstabandslaves)

	resMap := newDisksSummaryMap()
	resMap.rebuild(mblk)
	resMap.mergeFstabMap(ft)
	resMap.mergeMntMap(mounts)
	resMap.mergeUdevMap(udevDisks)

	timer := make(chan bool)
	if forceUpdate == nil {
		go func() {
			for {
				time.Sleep(time.Second * time.Duration(config.MonitoringFstabSec))
				timer <- true
			}
		}()
	}

	//fmt.Println("disks:")

	rch := make(chan DisksSummaryMap)

	copyResMap := func() DisksSummaryMap {
		dst := make(DisksSummaryMap)
		for k, v := range resMap {
			dst[k] = v
		}
		return dst
	}

	go func() {
		for {
			// make a copy for compare later
			oldMap := make(DisksSummaryMap, len(resMap))
			for k, v := range resMap {
				oldMap[k] = v
			}

			up := func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Recovered in up()", r)
					}
				}()

				mblk = fetchSysBlock("/sys/class/block")
				resMap.rebuild(mblk)
				udevDisks = newUdevMapInfo()
				// rescan ft
				resMap.mergeUdevMap(udevDisks)
				mounts = mapMntFile("/proc/mounts", udevDisks)
				resMap.mergeMntMap(mounts)
			}

			select {
			case rch <- copyResMap():
			case <-forceUpdate:
				up()
			case <-time.After(time.Millisecond * time.Duration(config.MonitoringProcmountMSec)):
				up()

			// udev monitoring
			// case event, ok := <-events:
			// 	if ok {
			// 		if devt, ok := event.Env["DEVTYPE"]; ok {
			// 			if devt == "disk" || devt == "partition" {
			// 				//fmt.Println(event.String())
			// 				name := strings.Split(event.Devpath, "/")
			// 				name = name[len(name)-1:]
			// 				//fmt.Println(event.Action, name, devt)

			// 				mblk = fetchSysBlock("/sys/class/block")
			// 				resMap.rebuild(mblk)
			// 				if !reflect.DeepEqual(resMap, oldMap) {
			// 					sendRes()
			// 				}
			// 			}
			// 		}
			// 	}

			// fstab monitoring (optional disabled)
			case <-timer:
				// for next scan mnt tick
				fstab = mapMntFile("/etc/fstab", udevDisks)
				fstabandslaves = mblk.exposeDevsSlaves(fstab.devPaths())
				ft = newFstabMap(fstab, fstabandslaves)
				resMap.mergeFstabMap(ft)

			case <-done:
				close(rch)
				//monitor.Close()
				return
			}
		}
	}()
	return rch
}
