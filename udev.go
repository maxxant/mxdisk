package mxdisk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// UdevadmInfo for mapping from:
// /sbin/udevadm info -q all -n /dev/sda1  (old and new Linux OS supports these format)
// newer Linux OS must works with: /sbin/udevadm info -p /sys/class/block/devXX
type UdevadmInfo struct {
	// E: key=val map
	ekv map[string]string
}

var onceFindUdevadm string

// NewUdevadmInfo new from param sysblk as: /dev/sdaXn
func NewUdevadmInfo(fdevname string) UdevadmInfo {
	if "" == onceFindUdevadm {
		path, err := exec.LookPath("udevadm")
		if err == nil {
			onceFindUdevadm = path
		} else {
			return UdevadmInfo{
				ekv: map[string]string{},
			}
		}
	}

	cmd := exec.Command(onceFindUdevadm, "info", "-q", "all", "-n", fdevname)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("udevadm failed:", err)
		return UdevadmInfo{
			ekv: map[string]string{},
		}
	}
	//fmt.Printf("out %s\n", out)

	res := UdevadmInfo{
		ekv: map[string]string{},
	}

	defer func() {
		// for recover in
		if x := recover(); x != nil {
			fmt.Println("panic in NewUdevadmInfo()", x)
		}
	}()

	res.parseOutput(udevadmRegexpE(), string(out))
	//fmt.Printf("out %s\n", res.ekv)
	return res
}

func udevadmRegexpE() *regexp.Regexp {
	return regexp.MustCompile("E: (\\w+)=(.+)")
}

func (p UdevadmInfo) parseOutput(rex *regexp.Regexp, str string) {
	data := rex.FindAllStringSubmatch(str, -1)

	for _, kv := range data {
		k := kv[1]
		v := kv[2]
		p.ekv[k] = v
	}
}

const (
	byUUID      = iota
	byLabel     = iota
	byPath      = iota
	byPartuuid  = iota
	byPartlabel = iota
)

// UdevInfo info from /dev/disk/by-xxx
type UdevInfo struct {
	UUID      string
	Label     string
	Path      string
	Partuuid  string
	Partlabel string
	phyParent string // physiscal device for partition. calculated value
}

// UdevMapInfo key = dev as /dev/sda1
type UdevMapInfo map[string]*UdevInfo

// newDisksByX for paths:
// - /dev/disk/by-uuid
// - /dev/disk/by-label
// - /dev/disk/by-path
// - /dev/disk/by-partuuid
// - /dev/disk/by-partlabel
// returns map [by-xxx] /dev/sdxN
// NOTE: not all OS supports path "by-label", "by-partlabel", "by-partuuid"
func newUdevMapInfo() UdevMapInfo {
	m := make(UdevMapInfo)
	m.fill4path("/dev/disk/by-uuid", byUUID)
	m.fill4path("/dev/disk/by-label", byLabel)
	m.fill4path("/dev/disk/by-path", byPath)
	m.fill4path("/dev/disk/by-partuuid", byPartuuid)
	m.fill4path("/dev/disk/by-partlabel", byPartlabel)
	m.buildPhy()
	return m
}

func (p UdevMapInfo) buildPhy() {
	markPhyParentForPathContains := func(tk string) {
		t := p[tk].Path
		if t != "" {
			for k, v := range p {
				if v.Path != t && strings.Contains(v.Path, t) {
					p[k].phyParent = tk
				}
			}
		}
	}

	for k := range p {
		markPhyParentForPathContains(k)
	}
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
			case byPartuuid:
				p[name].Partuuid = link
			case byPartlabel:
				p[name].Partlabel = link
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
		case byPartuuid:
			if v.Partuuid == needx {
				return k
			}
		case byPartlabel:
			if v.Partlabel == needx {
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
