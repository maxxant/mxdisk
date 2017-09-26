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

type disksByX struct {
	// key:val -> /dev/sda1 : uuid
	uuid  map[string]string
	label map[string]string
	path  map[string]string
}

// newDisksByX for paths:
// - /dev/disk/by-uuid
// - /dev/disk/by-label
// - /dev/disk/by-path
// TODO - /dev/disk/by-partuuid
// TODO - /dev/disk/by-partlabel
// returns map [by-xxx] /dev/sdxN
// NOTE: not all OS supports path "by-label", "by-partlabel", "by-partuuid"
func newDisksByX() *disksByX {
	fill4path := func(path string) map[string]string {
		mp := make(map[string]string)
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
				mp[name] = link
			}
			return err
		})
		return mp
	}

	return &disksByX{
		uuid:  fill4path("/dev/disk/by-uuid"),
		label: fill4path("/dev/disk/by-label"),
		path:  fill4path("/dev/disk/by-path"),
	}
}

func (p disksByX) getMap(byXFilter int) map[string]string {
	switch byXFilter {
	case byUUID:
		return p.uuid
	case byLabel:
		return p.label
	case byPath:
		return p.path
	}
	panic("undefined filter index")
}

func (p disksByX) findX(byXFilter int, dev string) string {
	mp := p.getMap(byXFilter)
	if v, ok := mp[dev]; ok {
		return v
	}
	return ""
}

func (p disksByX) findDevPath(byXFilter int, needx string) string {
	mp := p.getMap(byXFilter)
	for k, v := range mp {
		if v == needx {
			return k
		}
	}
	return ""
}

func (p disksByX) String() string {
	var s string
	for k, v := range p.uuid {
		s += fmt.Sprintf("uuid: %v : %v\n", k, v)
	}
	for k, v := range p.label {
		s += fmt.Sprintf("label: %v : %v\n", k, v)
	}
	for k, v := range p.path {
		s += fmt.Sprintf("path: %v : %v\n", k, v)
	}
	return s
}
