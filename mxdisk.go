package mxdisk

import (
	"fmt"
	"reflect"
	"time"
)

// WatchMounts return chan with removable storage info
// onlyUUID for mounted devs with UUID only for filtering /dev/loop, etc
func WatchMounts(done chan struct{}, config *Config, onlyUUID bool) chan MntMapDisks {
	mapDiskByX := newDisksByX()
	fmt.Println("diskBy:")
	fmt.Println(mapDiskByX)

	fstab := mapMntFile("/etc/fstab", mapDiskByX)
	fmt.Println("fstab:")
	fmt.Println(fstab)

	mounts := mapMntFile("/proc/mounts", mapDiskByX)
	fmt.Println("mounts:")
	fmt.Println(mounts)

	mblk := fetchSysBlock("/sys/block")
	fmt.Println("sysblock:")
	fmt.Println(mblk)

	fstabandslaves := mblk.exposeDevsSlaves(fstab.devPaths())
	fstabEx := mounts.devs4paths(fstabandslaves)
	fmt.Println("fstabEx:")
	fmt.Println(fstabEx)

	disks := getMntRemovableDisks(fstabEx, mounts, config)
	//fmt.Println(disks)

	fmt.Println("fstab-mounts mnt:")

	timer := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Second * time.Duration(config.MonitoringFstabSec))
			timer <- true
		}
	}()

	rch := make(chan MntMapDisks)
	go func() {
		rch <- disks
		for {
			select {
			case <-time.After(time.Second * time.Duration(config.MonitoringProcmountSec)):
				mapDiskByX = newDisksByX()
				mounts = mapMntFile("/proc/mounts", mapDiskByX)
				d := getMntRemovableDisks(fstabEx, mounts, config)
				if !reflect.DeepEqual(disks, d) {
					disks = d
					rch <- disks
				}

			case <-timer:
				fstab = mapMntFile("/etc/fstab", mapDiskByX)
				mblk = fetchSysBlock("/sys/block")
				fstabandslaves = mblk.exposeDevsSlaves(fstab.devPaths())
				fstabEx = mounts.devs4paths(fstabandslaves)

			case <-done:
				close(rch)
				return
			}
		}
	}()
	return rch
}
