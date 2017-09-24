package mxdisk

import (
	"fmt"
	"github.com/maxxant/go-fstab"
	"github.com/maxxant/udev"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// path:
//  /dev/disk/by-uuid
//  /dev/disk/by-label
//  /dev/disk/by-path
// returns map [by-xxx] /dev/sdxN
func disksByPathX(path string) map[string]string {
	mp := make(map[string]string)
	filepath.Walk(path, func(path string, inf os.FileInfo, err error) error {
		if inf.IsDir() {
			return nil
		}
		if (inf.Mode() & os.ModeSymlink) != 0 {
			base := filepath.Base(path)
			mp[base], _ = filepath.EvalSymlinks(path)
		}
		return err
	})
	return mp
}

// DiskInfo contains details mnt point, uuid, labels and others
type DiskInfo struct {
	DevPath  string
	MntPoint string
	UUID     string
	Label    string
	FsType   string
}

func mapMntFile(path string) map[string]DiskInfo {
	mpUUID := disksByPathX("/dev/disk/by-uuid")
	mpLabel := disksByPathX("/dev/disk/by-label")
	mp := make(map[string]DiskInfo)

	find4map := func(m map[string]string, needval string) string {
		for k, v := range m {
			if v == needval {
				return k
			}
		}
		return ""
	}

	if mnts, err := fstab.ParseFile(path); err == nil {
		for _, mnt := range mnts {

			fillDiskInfo := func(val string) {
				var fstype string
				if mnt.VfsType != "auto" {
					fstype = mnt.VfsType
				}

				mp[val] = DiskInfo{
					DevPath:  val,
					MntPoint: mnt.File,
					UUID:     find4map(mpUUID, val),
					Label:    find4map(mpLabel, val),
					FsType:   fstype,
				}
			}

			if mnt.SpecType() == fstab.Label || mnt.SpecType() == fstab.PartLabel {
				if val, ok := mpLabel[mnt.SpecValue()]; ok {
					fillDiskInfo(val)
				}
			} else if mnt.SpecType() == fstab.UUID || mnt.SpecType() == fstab.PartUUID {
				if val, ok := mpUUID[mnt.SpecValue()]; ok {
					fillDiskInfo(val)
				}
			} else if mnt.SpecType() == fstab.Path {
				if val, err := filepath.EvalSymlinks(mnt.SpecValue()); err == nil {
					fillDiskInfo(val)
				}
			}
		}
	}

	return mp
}

func getMntRemovableDisks(fstab map[string]DiskInfo, mounts map[string]DiskInfo) map[string]DiskInfo {
	res := make(map[string]DiskInfo)
	//fmt.Printf("mnts: %+v\n", mounts)

	findUUID := func(mp map[string]DiskInfo, uuid string) bool {
		for _, v := range mp {
			if v.UUID == uuid {
				return true
			}
		}
		return false
	}

	// check "/proc/mounts" records that not contains in "/etc/fstab" (by dev & uuid) and have non empty UUID (as block device)
	for k, v := range mounts {
		if _, ok := fstab[v.DevPath]; v.UUID != "" && !ok && !findUUID(fstab, v.UUID) {
			res[k] = v
		}
	}

	return res
}

// WatchMounts return chan with removable storage info
func WatchMounts(done chan struct{}) chan map[string]DiskInfo {
	rch := make(chan map[string]DiskInfo)
	fstab := mapMntFile("/etc/fstab")
	mounts := mapMntFile("/proc/mounts")
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

		if devt, ok := event.Env["DEVTYPE"]; ok && devt == "partition" {
			//fmt.Println(event.String())
			name := strings.Split(event.Devpath, "/")
			name = name[len(name)-1:]
			fmt.Println(event.Action, name, devt)
		}
	}
}
