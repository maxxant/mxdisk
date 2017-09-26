package mxdisk

import (
	"reflect"
	"sort"
	"testing"
)

func TestSysMapBlocks_exposeDevsSlaves(t *testing.T) {
	type args struct {
		devs []string
	}
	tests := []struct {
		name string
		p    SysMapBlocks
		args args
		want []string
	}{
		// test cases
		{
			name: "sda no slaves",
			p: SysMapBlocks{
				"/dev/sda": {DevPath: "/dev/sda"},
			},
			args: args{
				devs: []string{"/dev/sda"},
			},
			want: []string{
				"/dev/sda",
			},
		},
		{
			name: "md0 have sda1 & sdb1 slaves",
			p: SysMapBlocks{
				"/dev/md0": {
					DevPath: "/dev/md0",
					slaves: []string{
						"/dev/sda1",
						"/dev/sdb1",
					},
				},
			},
			args: args{
				devs: []string{"/dev/md0"},
			},
			want: []string{
				"/dev/md0",
				"/dev/sda1",
				"/dev/sdb1",
			},
		},
		{
			name: "recursion dm-1 have md0 slave and md0 have sda1 & sdb1 slaves",
			p: SysMapBlocks{
				"/dev/md0": {
					DevPath: "/dev/md0",
					slaves: []string{
						"/dev/sda1",
						"/dev/sdb1",
					},
				},
				"/dev/dm-1": {
					DevPath: "/dev/dm-1",
					slaves:  []string{"/dev/md0"},
				},
			},
			args: args{
				devs: []string{"/dev/dm-1"},
			},
			want: []string{
				"/dev/dm-1",
				"/dev/md0",
				"/dev/sda1",
				"/dev/sdb1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.p.exposeDevsSlaves(tt.args.devs)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SysMapBlocks.exposeDevsSlaves() = %v, want %v", got, tt.want)
			}
		})
	}
}
