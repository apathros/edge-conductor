/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"testing"
)

func Test_CheckCmdline(t *testing.T) {
	type args struct {
		cmdline string
		cmd     string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "no-force-download",
			args: args{
				cmdline: "",
				cmd:     "force-download",
			},
			want: false,
		},
		{
			name: "force-download",
			args: args{
				cmdline: "force-download\notherCmd\n",
				cmd:     "force-download",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckCmdline(tt.args.cmdline, tt.args.cmd); got != tt.want {
				t.Errorf("CheckCmdline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddCmdline(t *testing.T) {
	type args struct {
		s string
		t string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "return_test",
			args: args{
				s: "cmd1",
				t: "cmd2",
			},
			want: "cmd1\ncmd2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddCmdline(tt.args.s, tt.args.t); got != tt.want {
				t.Errorf("AddCmdline() = %v, want %v", got, tt.want)
			}
		})
	}
}
