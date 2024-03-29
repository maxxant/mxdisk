package mxdisk

import (
	"fmt"
	"path/filepath"

	"github.com/maxxant/go-fstab" // vendor fork from: github.com/deniswernert/go-fstab
)

// MntDiskInfo contains details mnt point, uuid, labels and others
type MntDiskInfo struct {
	MntPoints []string
	FsType    string
}

// MntMapDisks the map of mounted disks
type MntMapDisks map[string]MntDiskInfo

func (p MntMapDisks) String() string {
	var s string
	for k, v := range p {
		s += fmt.Sprintf("%v : %+v\n", k, v)
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

func mapMntFile(path string, mapby UdevMapInfo) MntMapDisks {
	mp := make(MntMapDisks)

	if mnts, err := fstab.ParseFile(path); err == nil {
		for _, mnt := range mnts {

			fillDiskInfo := func(val string) {
				var fstype string
				if mnt.VfsType != "auto" {
					fstype = mnt.VfsType
				}

				if v, ok := mp[val]; !ok {
					mp[val] = MntDiskInfo{
						MntPoints: []string{mnt.File},
						FsType:    fstype,
					}
				} else {
					// if _, ok := v.MntPoints[mnt.File]; !ok {
					// 	v.MntPoints[mnt.File] = struct{}{}
					// }
					v.MntPoints = append(v.MntPoints, mnt.File)
					mp[val] = v
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
