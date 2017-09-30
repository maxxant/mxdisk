package mxdisk

import "fmt"
import "sort"

// DiskSummary is info for disk and partition state
type DiskSummary struct {
	MntDiskInfo
	SysBlockInfo
	UdevInfo
}

// DisksSummaryMap map of disks
type DisksSummaryMap map[string]DiskSummary

func newDisksSummaryMap() DisksSummaryMap {
	return make(DisksSummaryMap)
}

func (p DisksSummaryMap) String() string {
	sl := sort.StringSlice(p.SliceKeys())
	sort.Sort(sl)

	var s string
	for _, k := range sl {
		s += fmt.Sprintf("%+v : %+v\n", k, p[k])
	}
	return s
}

// SliceKeys func convert map to slice
func (p DisksSummaryMap) SliceKeys() []string {
	da := make([]string, 0, len(p))
	for k := range p {
		da = append(da, k)
	}
	return da
}

// SliceValues func convert map to slice
func (p DisksSummaryMap) SliceValues() []DiskSummary {
	da := make([]DiskSummary, 0, len(p))
	for _, v := range p {
		da = append(da, v)
	}
	return da
}

func (p DisksSummaryMap) mergeSysMap(sys SysMapBlocks) {
	// add
	for k, v := range sys {
		if x, ok := p[k]; ok {
			x.SysBlockInfo = v
			p[k] = x
		} else {
			p[k] = DiskSummary{SysBlockInfo: v}
		}
	}

	// delete if the disk not present in /sys
	for k := range p {
		if _, ok := sys[k]; !ok {
			delete(p, k)
		}
	}
}

func (p DisksSummaryMap) mergeUdevMap(udev UdevMapInfo) {
	for k, v := range udev {
		if x, ok := p[k]; ok {
			x.UdevInfo = *v
			p[k] = x
		} else {
			//p[k] = DiskSummary{UdevInfo: *v}
			// disk not present in /sys
			panic("disk not present in /sys")
		}
	}
}

func (p DisksSummaryMap) mergeMntMap(mnt MntMapDisks) {
	for k, v := range mnt {
		if x, ok := p[k]; ok {
			x.MntDiskInfo = v
			p[k] = x
		} else {
			//p[k] = DiskSummary{MntDiskInfo: v}
			// disk not present in /sys
			panic("disk not present in /sys")
		}
	}
}

// func (p DisksSummaryMap) minusFstab(fstab MntMapDisks, config *Config) {
// 	//res := make(MntMapDisks)
// 	findUUID := func(mp MntMapDisks, uuid string) bool {
// 		for _, v := range mp {
// 			if v.UUID == uuid {
// 				return true
// 			}
// 		}
// 		return false
// 	}

// 	// check "/proc/mounts" records that not contains in "/etc/fstab" (by dev & UUID) and fstab's RAID slaves)
// 	// and optional have non empty UUID as block device (for example /dev/loop is not have UUIDs and will be filtered out)
// 	for k, v := range p {
// 		if _, ok := fstab[k]; (!config.OnlyUUIDMountedDisks || v.UUID != "") && !ok && !findUUID(fstab, v.UUID) {
// 			//res[k] = v
// 		} else {
// 			delete(p, k)
// 		}
// 	}
// }
