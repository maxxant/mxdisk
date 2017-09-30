package mxdisk

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	byUUID  = iota
	byLabel = iota
	byPath  = iota
)

// UdevInfo info from /dev/disk/by-xxx
type UdevInfo struct {
	UUID  string
	Label string
	Path  string
}

// UdevMapInfo key = dev as /dev/sda1
type UdevMapInfo map[string]*UdevInfo

// newDisksByX for paths:
// - /dev/disk/by-uuid
// - /dev/disk/by-label
// - /dev/disk/by-path
// TODO - /dev/disk/by-partuuid
// TODO - /dev/disk/by-partlabel
// returns map [by-xxx] /dev/sdxN
// NOTE: not all OS supports path "by-label", "by-partlabel", "by-partuuid"
func newUdevMapInfo() UdevMapInfo {
	m := make(UdevMapInfo)
	m.fill4path("/dev/disk/by-uuid", byUUID)
	m.fill4path("/dev/disk/by-label", byLabel)
	m.fill4path("/dev/disk/by-path", byPath)
	return m
}

func (p UdevMapInfo) fill4path(path string, byX int) {
	filepath.Walk(path, func(path string, inf os.FileInfo, err error) error {
		if err != nil {
			return err // if path is not exists
		}
		if inf.IsDir() {
			return nil
		}
		if (inf.Mode() & os.ModeSymlink) != 0 {
			link := filepath.Base(path)
			name, _ := filepath.EvalSymlinks(path)

			if _, ok := p[name]; !ok {
				p[name] = &UdevInfo{}
			}
			switch byX {
			case byUUID:
				p[name].UUID = link
			case byLabel:
				p[name].Label = link
			case byPath:
				p[name].Path = link
			}
		}
		return err
	})
}

func (p UdevMapInfo) findDevPath(byXFilter int, needx string) string {
	for k, v := range p {
		switch byXFilter {
		case byUUID:
			if v.UUID == needx {
				return k
			}
		case byLabel:
			if v.Label == needx {
				return k
			}
		case byPath:
			if v.Path == needx {
				return k
			}
		}
	}
	return ""
}

func (p UdevMapInfo) String() string {
	var s string
	for k, v := range p {
		s += fmt.Sprintf("%v : %+v\n", k, v)
	}
	return s
}
