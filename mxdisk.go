package mxdisk

import (
	"reflect"
	"time"
)

// WatchMounts return chan with removable storage info
func WatchMounts(done chan struct{}) chan map[string]DiskInfo {
	rch := make(chan map[string]DiskInfo)
	fstab := mapMntFile("/etc/fstab")
	mounts := mapMntFile("/proc/mounts")
	fetchSysBlock("/sys/block")
	disks := getMntRemovableDisks(fstab, mounts)

	timer := make(chan bool)
	go func() {
		for {
			time.Sleep(time.Second * 15)
			timer <- true
		}
	}()

	go func() {
		rch <- disks
		for {
			select {
			case <-time.After(time.Second * 1):
				mounts = mapMntFile("/proc/mounts")
				d := getMntRemovableDisks(fstab, mounts)
				if !reflect.DeepEqual(disks, d) {
					disks = d
					rch <- disks
				}

			case <-timer:
				fstab = mapMntFile("/etc/fstab")

			case <-done:
				close(rch)
				return
			}
		}
	}()
	return rch
}
