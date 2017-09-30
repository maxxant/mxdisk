package mxdisk

import (
	"fmt"
	"sort"
)

// Partition info
type Partition struct {
	MntDiskInfo
	UdevInfo
}

// Disk with partitions map
type Disk struct {
	Fstab
	Virtual bool
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

			if v.FstabMention {
				d.FstabMention = true
			}

			// case for disk and partition in one as ISO-fs devices v.UUID != ""
			// and for virtual partitions as /dev/loop
			if v.UUID != "" || len(v.MntPoints) > 0 {
				d.Part[k] = Partition{
					MntDiskInfo: v.MntDiskInfo,
					UdevInfo:    v.UdevInfo,
				}
				if v.FstabMention {
					d.FstabMention = true
				}
			}
			if v.Path == "" {
				d.Virtual = true
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
					}
					if sv.FstabMention {
						v.FstabMention = true
					}
				}
			}
			mp[k] = v
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
