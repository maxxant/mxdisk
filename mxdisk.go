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
	*MntDiskInfo
	*SysBlockInfo
}

// DisksArr map of disks
type DisksArr map[string]Disk

func newDisksArr() DisksArr {
	return make(DisksArr)
}

func (p DisksArr) String() string {
	var s string
	for k, v := range p {
		s += fmt.Sprintf("%+v : %+v\n", k, v)
	}
	return s
}

// Slice func convert map to slice
func (p DisksArr) Slice() []Disk {
	da := make([]Disk, 0, len(p))
	for _, v := range p {
		da = append(da, v)
	}
	return da
}

func (p DisksArr) mergeMntMap(mnt MntMapDisks) {
	if p == nil {
		return
	}

	for k, v := range mnt {
		p[k].MntDiskInfo = &v
	}
}

func (p DisksArr) mergeSysMap(sys SysMapBlocks) {
	if p == nil {
		return
	}

	for k, v := range sys {
		p[k] = Disk{SysBlockInfo: &v}
	}
}

func (p DisksArr) minusFstab(fstab MntMapDisks, config *Config) {
	//res := make(MntMapDisks)
	findUUID := func(mp MntMapDisks, uuid string) bool {
		for _, v := range mp {
			if v.UUID == uuid {
				return true
			}
		}
		return false
	}

	// check "/proc/mounts" records that not contains in "/etc/fstab" (by dev & UUID) and fstab's RAID slaves)
	// and optional have non empty UUID as block device (for example /dev/loop is not have UUIDs and will be filtered out)
	for k, v := range p {
		if _, ok := fstab[k]; (!config.OnlyUUIDMountedDisks || v.UUID != "") && !ok && !findUUID(fstab, v.UUID) {
			//res[k] = v
		} else {
			delete(p, k)
		}
	}
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

	mblk := fetchSysBlock("/sys/class/block")
	fmt.Println("sysblock:")
	fmt.Println(mblk)

	fstabandslaves := mblk.exposeDevsSlaves(fstab.devPaths())
	fstabEx := mounts.devs4paths(fstabandslaves)
	fmt.Println("fstabEx:")
	fmt.Println(fstabEx)

	//mdisks := getMntRemovableDisks(fstabEx, mounts, config)
	resMap := newDisksArr()
	resMap.mergeMntMap(mounts)
	resMap.mergeSysMap(mblk)
	resMap.minusFstab(fstabEx, config)

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

	fmt.Println("nofstab disks:")

	rch := make(chan DisksArr)
	go func() {
		rch <- resMap
		for {

			// make a copy for compare later
			oldMap := make(DisksArr, len(resMap))
			for k, v := range resMap {
				oldMap[k] = v
			}

			select {
			// mnt monitoring
			case <-time.After(time.Second * time.Duration(config.MonitoringProcmountSec)):
				mapDiskByX = newDisksByX()
				mounts = mapMntFile("/proc/mounts", mapDiskByX)
				resMap.mergeMntMap(mounts)
				resMap.minusFstab(fstabEx, config)
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

							mblk = fetchSysBlock("/sys/block")
							resMap.mergeSysMap(mblk)
							resMap.minusFstab(fstabEx, config)
							rch <- resMap
						}
					}
				}

			// fstab monitoring (optional disabled)
			case <-timer:
				// for next scan mnt tick
				fstab = mapMntFile("/etc/fstab", mapDiskByX)
				fstabandslaves = mblk.exposeDevsSlaves(fstab.devPaths())
				fstabEx = mounts.devs4paths(fstabandslaves)
				//resMap.minusFstab(fstabEx, config)
				//rch <- resMap

				// reload /sys/block without udev events and take effect in mnt monitoring tick
				mblk = fetchSysBlock("/sys/block")
				resMap.mergeSysMap(mblk)

			case <-done:
				close(rch)
				monitor.Close()
				return
			}
		}
	}()
	return rch
}
