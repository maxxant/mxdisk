package mxdisk

import (
	"fmt"
	"path/filepath"

	"github.com/maxxant/go-fstab" // vendor fork from: github.com/deniswernert/go-fstab
)

// MntDiskInfo contains details mnt point, uuid, labels and others
type MntDiskInfo struct {
	DevPath  string
	MntPoint string
	UUID     string
	Label    string
	FsType   string
}

// MntMapDisks the map of mounted disks
type MntMapDisks map[string]MntDiskInfo

func (p MntMapDisks) String() string {
	var s string
	for _, v := range p {
		s += fmt.Sprintf("%+v\n", v)
	}
	return s
}

func (p MntMapDisks) devPaths() []string {
	var s []string
	for k := range p {
		s = append(s, k)
	}
	return s
}

func (p MntMapDisks) devs4paths(paths []string) MntMapDisks {
	mp := make(MntMapDisks)
	for _, v := range paths {
		if _, ok := p[v]; ok {
			mp[v] = p[v]
		} else {
			// spec case: record dev name only for deep slaves disks,
			// because any others info it is not available.
			// (example : dm-1 with slave RAID md1 and sda1 & sdb1 slaves)
			mp[v] = MntDiskInfo{
				DevPath: v,
			}
		}
	}
	return mp
}

func mapMntFile(path string, mapby *disksByX) MntMapDisks {
	mp := make(MntMapDisks)

	if mnts, err := fstab.ParseFile(path); err == nil {
		for _, mnt := range mnts {

			fillDiskInfo := func(val string) {
				var fstype string
				if mnt.VfsType != "auto" {
					fstype = mnt.VfsType
				}

				mp[val] = MntDiskInfo{
					DevPath:  val,
					MntPoint: mnt.File,
					UUID:     mapby.findX(byUUID, val),
					Label:    mapby.findX(byLabel, val),
					FsType:   fstype,
				}
			}

			if mnt.SpecType() == fstab.Label || mnt.SpecType() == fstab.PartLabel {
				if val := mapby.findDevPath(byLabel, mnt.SpecValue()); val != "" {
					fillDiskInfo(val)
				}
			} else if mnt.SpecType() == fstab.UUID || mnt.SpecType() == fstab.PartUUID {
				if val := mapby.findDevPath(byUUID, mnt.SpecValue()); val != "" {
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

func getMntRemovableDisks(fstab MntMapDisks, mounts MntMapDisks, config *Config) MntMapDisks {
	res := make(MntMapDisks)
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
	for k, v := range mounts {
		if _, ok := fstab[v.DevPath]; (!config.OnlyUUIDMountedDisks || v.UUID != "") && !ok && !findUUID(fstab, v.UUID) {
			res[k] = v
		}
	}

	return res
}
