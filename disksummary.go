package mxdisk

import "fmt"
import "sort"

// DiskSummary is info for disk and partition state
type DiskSummary struct {
	MntDiskInfo
	SysBlockInfo
	UdevInfo
	Fstab
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

func (p DisksSummaryMap) rebuild(sys SysMapBlocks) {
	// add /sys
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

func (p DisksSummaryMap) mergeFstabMap(fstab FstabMap) {
	for k, v := range p {
		if x, ok := fstab[k]; ok {
			v.Fstab = x
			p[k] = v
		} else {
			v.Fstab = Fstab{}
			p[k] = v
		}
	}
}

func (p DisksSummaryMap) mergeUdevMap(udev UdevMapInfo) {
	for k, v := range p {
		if x, ok := udev[k]; ok {
			v.UdevInfo = *x
			p[k] = v
		} else {
			v.UdevInfo = UdevInfo{}
			p[k] = v
		}
	}
}

func (p DisksSummaryMap) mergeMntMap(mnt MntMapDisks) {
	for k, v := range p {
		if x, ok := mnt[k]; ok {
			v.MntDiskInfo = x
			p[k] = v
		} else {
			v.MntDiskInfo = MntDiskInfo{}
			p[k] = v
		}
	}
}
