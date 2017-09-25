package mxdisk

import (
	"fmt"
	"github.com/maxxant/go-fstab" // vendor fork from: github.com/deniswernert/go-fstab
	"os"
	"path/filepath"
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

// path:
//  /dev/disk/by-uuid
//  /dev/disk/by-label
//  /dev/disk/by-path
// returns map [by-xxx] /dev/sdxN
// NOTE: not all OS supports path "by-label"
func disksByPathX(path string) map[string]string {
	mp := make(map[string]string)
	filepath.Walk(path, func(path string, inf os.FileInfo, err error) error {
		if err != nil {
			return err // if path is not exists
		}
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

type disksByX struct {
	uuid  map[string]string
	label map[string]string
	//path map[string]string
}

func newDisksByX() *disksByX {
	return &disksByX{
		uuid:  disksByPathX("/dev/disk/by-uuid"),
		label: disksByPathX("/dev/disk/by-label"),
		//path : disksByPathX("/dev/disk/by-path"),
	}
}

func mapMntFile(path string, mapby *disksByX) MntMapDisks {
	mp := make(MntMapDisks)

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

				mp[val] = MntDiskInfo{
					DevPath:  val,
					MntPoint: mnt.File,
					UUID:     find4map(mapby.uuid, val),
					Label:    find4map(mapby.label, val),
					FsType:   fstype,
				}
			}

			if mnt.SpecType() == fstab.Label || mnt.SpecType() == fstab.PartLabel {
				if val, ok := mapby.label[mnt.SpecValue()]; ok {
					fillDiskInfo(val)
				}
			} else if mnt.SpecType() == fstab.UUID || mnt.SpecType() == fstab.PartUUID {
				if val, ok := mapby.uuid[mnt.SpecValue()]; ok {
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
