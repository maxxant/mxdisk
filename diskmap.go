package mxdisk

import (
	"fmt"
	"sort"
)

// Partition info
type Partition struct {
	MntDiskInfo
	UdevInfo
	Fstab
}

// Disk with partitions map
type Disk struct {
	SysBlockInfo
	Part map[string]Partition
}

// DiskMap disks tree
type DiskMap map[string]Disk

// NewDiskMap make DiskMap from DisksSummaryMap
func NewDiskMap(sum DisksSummaryMap) DiskMap {
	mp := make(DiskMap)

	for k, v := range sum {
		if v.phyParent == "" {
			d := Disk{
				SysBlockInfo: v.SysBlockInfo,
				Part:         map[string]Partition{},
			}

			// case for disk and partition in one (iso fs devices)
			if v.UUID != "" {
				d.Part[k] = Partition{
					MntDiskInfo: v.MntDiskInfo,
					UdevInfo:    v.UdevInfo,
					Fstab:       v.Fstab,
				}
			}
			mp[k] = d
		}
	}

	// fill childs partitions

	for k, v := range mp {
		// ignore 2in1 devices, its already filled in previous step
		if len(v.Part) == 0 {
			for sk, sv := range sum {
				if sv.phyParent == k {
					v.Part[sk] = Partition{
						MntDiskInfo: sv.MntDiskInfo,
						UdevInfo:    sv.UdevInfo,
						Fstab:       sv.Fstab,
					}
				}
			}
		}
	}

	return mp
}

func (p DiskMap) String() string {
	sl := sort.StringSlice(p.SliceKeys())
	sort.Sort(sl)

	var s string
	for _, k := range sl {
		s += fmt.Sprintf("%+v : %+v\n", k, p[k])
	}
	return s
}

// SliceKeys func convert map to slice
func (p DiskMap) SliceKeys() []string {
	da := make([]string, 0, len(p))
	for k := range p {
		da = append(da, k)
	}
	return da
}
