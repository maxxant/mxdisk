package mxdisk

import "fmt"

// DiskSummary is info for disk and partition state
type DiskSummary struct {
	MntDiskInfo
	SysBlockInfo
	//UdevInfo
}

// DisksSummaryMap map of disks
type DisksSummaryMap map[string]DiskSummary

func newDisksSummaryMap() DisksSummaryMap {
	return make(DisksSummaryMap)
}

func (p DisksSummaryMap) String() string {
	var s string
	for k, v := range p {
		s += fmt.Sprintf("%+v : %+v\n", k, v)
	}
	return s
}

// Slice func convert map to slice
func (p DisksSummaryMap) Slice() []DiskSummary {
	da := make([]DiskSummary, 0, len(p))
	for _, v := range p {
		da = append(da, v)
	}
	return da
}

func (p DisksSummaryMap) mergeMntMap(mnt MntMapDisks) {
	for k, v := range mnt {
		if x, ok := p[k]; ok {
			x.MntDiskInfo = v
			p[k] = x
		} else {
			p[k] = DiskSummary{MntDiskInfo: v}
		}
	}
}

func (p DisksSummaryMap) mergeSysMap(sys SysMapBlocks) {
	for k, v := range sys {
		if x, ok := p[k]; ok {
			x.SysBlockInfo = v
			p[k] = x
		} else {
			p[k] = DiskSummary{SysBlockInfo: v}
		}
	}
}

func (p DisksSummaryMap) minusFstab(fstab MntMapDisks, config *Config) {
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
