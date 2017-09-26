package mxdisk

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/maxxant/udev" // vendor fork from: github.com/deniswernert/udev
)

// Disk is info for disk and partition state
type Disk struct {
	MntDiskInfo
	SysBlockInfo
}

// DisksArr slice of disks
type DisksArr []Disk

func (p DisksArr) String() {
	for v := range p {
		fmt.Println(v)
	}
}

func mergeDisksMap2Slise(mnt MntMapDisks, sys SysMapBlocks) DisksArr {
	dset := make(map[string]Disk)

	for k, v := range mnt {
		dset[k] = Disk{MntDiskInfo: v}
	}

	for k, v := range sys {
		dset[k] = Disk{SysBlockInfo: v}
	}

	da := make(DisksArr, 0, len(dset))
	for _, v := range dset {
		da = append(da, v)
	}
	return da
}

// Watch return chan with removable storage info
// onlyUUID for mounted devs with UUID only for filtering /dev/loop, etc
func Watch(done chan struct{}, config *Config, onlyUUID bool) chan DisksArr {
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

	mdisks := getMntRemovableDisks(fstabEx, mounts, config)
	//fmt.Println(disks)

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

	fmt.Println("fstab-mounts mnt:")

	rch := make(chan DisksArr)
	go func() {
		disks := mergeDisksMap2Slise(mdisks, mblk)
		rch <- disks
		for {
			select {
			case <-time.After(time.Second * time.Duration(config.MonitoringProcmountSec)):
				mapDiskByX = newDisksByX()
				mounts = mapMntFile("/proc/mounts", mapDiskByX)
				d := getMntRemovableDisks(fstabEx, mounts, config)
				if !reflect.DeepEqual(mdisks, d) {
					mdisks = d
					disks = mergeDisksMap2Slise(mdisks, mblk)
					rch <- disks
				}

			case event, ok := <-events:
				if ok {
					if devt, ok := event.Env["DEVTYPE"]; ok {
						if devt == "disk" || devt == "partition" {
							//fmt.Println(event.String())
							name := strings.Split(event.Devpath, "/")
							name = name[len(name)-1:]
							fmt.Println(event.Action, name, devt)

							mblk = fetchSysBlock("/sys/block")
							disks = mergeDisksMap2Slise(mdisks, mblk)
							rch <- disks
						}
					}
				} else {
					// TODO rm
					panic("ev")
				}

			case <-timer:
				mblk = fetchSysBlock("/sys/block")
				disks = mergeDisksMap2Slise(mdisks, mblk)
				//rch <- disks
				// for next scan mnt tick
				fstab = mapMntFile("/etc/fstab", mapDiskByX)
				fstabandslaves = mblk.exposeDevsSlaves(fstab.devPaths())
				fstabEx = mounts.devs4paths(fstabandslaves)

			case <-done:
				close(rch)
				monitor.Close()
				return
			}
		}
	}()
	return rch
}
