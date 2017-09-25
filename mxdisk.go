package mxdisk

import (
	"fmt"
	"reflect"
	"time"
)

// WatchMounts return chan with removable storage info
func WatchMounts(done chan struct{}) chan MntMapDisks {
	rch := make(chan MntMapDisks)
	fstab := mapMntFile("/etc/fstab")
	fmt.Println("fstab:")
	fmt.Println(fstab)
	mounts := mapMntFile("/proc/mounts")
	mblk := fetchSysBlock("/sys/block")
	fmt.Println("sysblock:")
	fmt.Println(mblk)
	disks := getMntRemovableDisks(fstab, mounts)
	//fmt.Println(disks)

	fmt.Println("fstab-mounts mnt:")
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
