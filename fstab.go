package mxdisk

import (
	"fmt"
)

// Fstab info that disk fstab presents
type Fstab struct {
	Clearly  bool
	Indirect bool
}

// FstabMap devkey:info map
type FstabMap map[string]Fstab

func newFstabMap(mnt MntMapDisks, allpaths []string) FstabMap {
	mp := make(FstabMap)
	for k := range mnt {
		mp[k] = Fstab{Clearly: true}
	}

	for _, v := range allpaths {
		if _, ok := mp[v]; !ok {
			mp[v] = Fstab{Indirect: true}
		}
	}
	return mp
}

func (p FstabMap) String() string {
	var s string
	for k, v := range p {
		s += fmt.Sprintf("%v : %+v\n", k, v)
	}
	return s
}
