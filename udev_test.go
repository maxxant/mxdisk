package mxdisk

import (
	"reflect"
	"regexp"
	"testing"
)

func TestUdevadmInfo_parseOutput(t *testing.T) {
	type fields struct {
		ekv map[string]string
	}
	type args struct {
		rex *regexp.Regexp
		str string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   fields
	}{
		{
			name: "udevadm KV parser",
			fields: fields{
				ekv: map[string]string{},
			},
			args: args{
				rex: udevadmRegexpE(),
				str: `N: sdd5
				E: DEVLINKS=/dev/disk/by-path/pci-0000:00:1f.2-ata-4-part5 /dev/disk/by-uuid/xxxx-xxxx-xxx
				E: ID_BUS=ata
				E: ID_FS_TYPE=ext4`,
			},
			want: fields{
				ekv: map[string]string{
					"DEVLINKS":   "/dev/disk/by-path/pci-0000:00:1f.2-ata-4-part5 /dev/disk/by-uuid/xxxx-xxxx-xxx", // double values
					"ID_FS_TYPE": "ext4",
					"ID_BUS":     "ata",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := UdevadmInfo{
				ekv: tt.fields.ekv,
			}
			p.parseOutput(tt.args.rex, tt.args.str)
			if !reflect.DeepEqual(tt.fields, tt.want) {
				t.Errorf("UdevadmInfo.parseOutput() = %v, want %v", tt.fields, tt.want)
			}
		})
	}
}
