/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package eputils

import (
	pluginapi "ep/pkg/api/plugins"
	fakesftp_mock "ep/pkg/eputils/test/fakesftp/mock"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	sshd "github.com/gliderlabs/ssh"
	gomock "github.com/golang/mock/gomock"
	sftp "github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	mpatch "github.com/undefinedlabs/go-mpatch"
	"golang.org/x/crypto/ssh"
)

const (
	sshPort int64  = 2222
	sshHost string = "127.0.0.1"
)

var nodeWithKey = pluginapi.Node{
	SSHPasswd: "123456",
	SSHPort:   sshPort,
	IP:        sshHost,
	User:      "root",
}

var nodeWithWrongPort = pluginapi.Node{
	SSHPasswd: "123456",
	SSHPort:   2223,
	IP:        sshHost,
	User:      "root",
}
var nodeWithWrongKey = pluginapi.Node{
	SSHKey:    "abcde",
	SSHPasswd: "123456",
	SSHPort:   sshPort,
	IP:        sshHost,
	User:      "root",
}

var nodeNoKey = pluginapi.Node{
	SSHPasswd: "123456",
	SSHPort:   sshPort,
	IP:        sshHost,
	User:      "root",
}

var (
	errCmd       = errors.New("cmd err")
	errSess      = errors.New("Session err")
	errStdIn     = errors.New("Stdin err")
	errSh        = errors.New("Shell err")
	errCl        = errors.New("Client err")
	errWait      = errors.New("Wait err")
	errFmt       = errors.New("fmt fake err")
	errCase      = errors.New("case err")
	errNoFile    = errors.New("CopyRemoteFileToLocalFile: not a file")
	errFileOrDir = errors.New("no such file or directory")
	errClnt      = errors.New("client err")
	errSession   = errors.New("session err")
	errConRefuse = fmt.Errorf("connect: connection refused")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()

	if err != nil {
		t.Fatal(err)
	}
}

func patchCopyLocalFileToRemoteFile(t *testing.T, retErr error) {
	var patch *mpatch.Patch
	var err error
	patch, err = mpatch.PatchMethod(CopyLocalFileToRemoteFile, func(addr string, cfg *ssh.ClientConfig, localPath, remotePath string) error {
		unpatch(t, patch)
		return retErr
	})
	if err != nil {
		t.Fatal(err)
	}
}

func patchRunRemoteCMD(t *testing.T, retErr error) {
	var patch *mpatch.Patch
	var err error
	patch, err = mpatch.PatchMethod(RunRemoteCMD, func(addr string, cfg *ssh.ClientConfig, cmd string) error {
		unpatch(t, patch)
		return retErr
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenSSHConfig(t *testing.T) {
	t.Log("Start Test GenSSHConfig")
	cases := []struct {
		name        string
		node        *pluginapi.Node
		retError    error
		expectError bool
	}{
		{
			"WithSSHKeyNode",
			&nodeWithKey,
			nil,
			false,
		},

		{
			"WithWrongSSHKeyNode",
			&nodeWithWrongKey,
			nil,
			true,
		},

		{
			"NoSSHKeyNode",
			&nodeNoKey,
			nil,
			false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			_, err := GenSSHConfig(tc.node)
			if tc.expectError == false && err != nil {
				t.Error(err)
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}
}

func TestRunRemoteCMD(t *testing.T) {
	t.Log("Start Test RunRemoteCMD")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cases := []struct {
		name       string
		node       *pluginapi.Node
		cmd        string
		clienterr  error
		sessionerr error
		cmderr     error
	}{
		{
			"Normal Case",
			&nodeWithKey,
			"echo hi",
			nil,
			nil,
			nil,
		},
		{
			"Client Error",
			&nodeWithWrongPort,
			"echo hi",
			errClnt,
			nil,
			nil,
		},
		{
			"Session Error",
			&nodeWithKey,
			"echo hi",
			nil,
			errSession,
			nil,
		},
		{
			"Cmd Error",
			&nodeWithKey,
			"echo hi",
			nil,
			nil,
			errCmd,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			var c *ssh.Client
			if tc.sessionerr != nil {
				patch, err := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(c),
					"NewSession",
					func(c *ssh.Client) (*ssh.Session, error) {
						return nil, tc.sessionerr
					})
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch)
			}
			var s *ssh.Session
			patch2, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(s),
				"Run",
				func(s *ssh.Session, cmd string) error {
					return tc.cmderr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch2)
			err = RunRemoteCMD(addr, cfg, tc.cmd)
			if err != nil {
				if tc.clienterr != nil || tc.sessionerr != nil || tc.cmderr != nil {
					t.Logf("Test Case %s Pass", tc.name)
				} else {
					t.Logf("Test Case %s Fail", tc.name)
				}
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}

}

func TestRunRemoteMultiCMD(t *testing.T) {
	t.Log("Start Test RunRemoteMultiCMD")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cases := []struct {
		name        string
		node        *pluginapi.Node
		cmd         []string
		clienterr   error
		sessionerr  error
		stdinpiperr error
		shellerr    error
		waiterr     error
		fmtmock     bool
	}{
		{
			"Normal Case",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			nil,
			nil,
			false,
		},
		{
			"Fprintf Case",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			nil,
			nil,
			true,
		},
		{
			"Client Error",
			&nodeWithWrongPort,
			[]string{"echo hi"},
			errCl,
			nil,
			nil,
			nil,
			nil,
			false,
		},
		{
			"Session Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			errSess,
			nil,
			nil,
			nil,
			false,
		},
		{
			"std Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			errStdIn,
			nil,
			nil,
			false,
		},
		{
			"Shell Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			errSh,
			nil,
			false,
		},
		{
			"Wait Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			nil,
			errWait,
			false,
		},
	}

	for item, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			var c *ssh.Client
			if tc.sessionerr != nil {
				patch, err := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(c),
					"NewSession",
					func(c *ssh.Client) (*ssh.Session, error) {
						return nil, tc.sessionerr
					})
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch)
			}
			if tc.fmtmock {
				patch2, err := mpatch.PatchMethod(fmt.Fprintf, func(w io.Writer, format string, a ...interface{}) (n int, err error) {
					return 0, errFmt
				})
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch2)
			}
			var s *ssh.Session
			if tc.stdinpiperr != nil {
				patch3, err := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(s),
					"StdinPipe",
					func(s *ssh.Session) (io.WriteCloser, error) {
						return nil, tc.stdinpiperr
					})

				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch3)
			}
			patch4, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(s),
				"Shell",
				func(s *ssh.Session) error {
					return tc.shellerr
				})

			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch4)
			patch5, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(s),
				"Wait",
				func(s *ssh.Session) error {
					return tc.waiterr
				})

			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch5)
			err = RunRemoteMultiCMD(addr, cfg, tc.cmd)
			if err != nil {
				if item > 1 {
					t.Logf("Test Case %s Pass", tc.name)
				} else {
					t.Logf("Test Case %s Fail", tc.name)
				}
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}

}

func TestRunRemoteNodeMultiCMD(t *testing.T) {
	t.Log("Start Test RunRemoteNodeMultiCMD")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cases := []struct {
		name        string
		node        *pluginapi.Node
		cmd         []string
		clienterr   error
		sessionerr  error
		stdinpiperr error
		shellerr    error
		waiterr     error
		fmtmock     bool
	}{
		{
			"Normal Case",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			nil,
			nil,
			false,
		},
		{
			"Fprintf Case",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			nil,
			nil,
			true,
		},
		{
			"Client Error",
			&nodeWithWrongPort,
			[]string{"echo hi"},
			errCl,
			nil,
			nil,
			nil,
			nil,
			false,
		},
		{
			"Session Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			errSess,
			nil,
			nil,
			nil,
			false,
		},
		{
			"std Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			errStdIn,
			nil,
			nil,
			false,
		},
		{
			"Shell Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			errSh,
			nil,
			false,
		},
		{
			"Wait Error",
			&nodeWithKey,
			[]string{"echo hi"},
			nil,
			nil,
			nil,
			nil,
			errWait,
			false,
		},
	}

	for item, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			var c *ssh.Client
			if tc.sessionerr != nil {
				patch, err := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(c),
					"NewSession",
					func(c *ssh.Client) (*ssh.Session, error) {
						return nil, tc.sessionerr
					})
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch)
			}
			if tc.fmtmock {
				patch2, err := mpatch.PatchMethod(fmt.Fprintf, func(w io.Writer, format string, a ...interface{}) (n int, err error) {
					return 0, errFmt
				})
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch2)
			}
			var s *ssh.Session
			if tc.stdinpiperr != nil {
				patch3, err := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(s),
					"StdinPipe",
					func(s *ssh.Session) (io.WriteCloser, error) {
						return nil, tc.stdinpiperr
					})

				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch3)
			}
			patch4, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(s),
				"Shell",
				func(s *ssh.Session) error {
					return tc.shellerr
				})

			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch4)
			patch5, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(s),
				"Wait",
				func(s *ssh.Session) error {
					return tc.waiterr
				})

			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch5)
			err = RunRemoteNodeMultiCMD(tc.node, tc.cmd)
			if err != nil {
				if item > 1 {
					t.Logf("Test Case %s Pass", tc.name)
				} else {
					t.Logf("Test Case %s Fail", tc.name)
				}
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}

}

func TestWriteRemoteFile(t *testing.T) {
	t.Log("Start Test WriteRemoteFile")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	errresult := errCase
	cases := []struct {
		name      string
		node      *pluginapi.Node
		content   string
		path      string
		clienterr error
		mkdirerr  error
		createerr error
		closeerr  error
		chmoderr  error
		writeerr  error
	}{
		{"Normal Case",
			&nodeWithKey,
			"test",
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		},
		{"Dial Error Case",
			&nodeWithWrongPort,
			"test",
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		},
		{"Client Error Case",
			&nodeWithKey,
			"test",
			"./test/data/test",
			errresult,
			nil,
			nil,
			nil,
			nil,
			nil,
		},

		{"Mkdir Fail Case",
			&nodeWithKey,
			"test",
			"./test/data/test",
			nil,
			errresult,
			nil,
			nil,
			nil,
			nil,
		},
		{"Create Fail Case",
			&nodeWithKey,
			"test",
			"./test/data/test",
			nil,
			nil,
			errresult,
			nil,
			nil,
			nil,
		},
		{"Close Fail Case",
			&nodeWithKey,
			"test",
			"./test/data/test",
			nil,
			nil,
			nil,
			errresult,
			nil,
			nil,
		},
		{"Chmod Fail Case",
			&nodeWithKey,
			"test",
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			errresult,
			nil,
		},
		{"Write Fail Case",
			&nodeWithKey,
			"test",
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			nil,
			errresult,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			var c *sftp.Client
			var f *sftp.File
			mocksftpInterface := fakesftp_mock.NewMockSftpClientInterface(ctrl)
			patchc, err := mpatch.PatchMethod(sftp.NewClient, mocksftpInterface.NewClient)
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patchc)

			mocksftpInterface.EXPECT().NewClient(gomock.Any()).AnyTimes().Return(&sftp.Client{}, tc.clienterr)
			patch, _ := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(c),
				"MkdirAll",
				func(c *sftp.Client, path string) error {
					return tc.mkdirerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch)
			patch2, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(c),
				"Create",
				func(c *sftp.Client, path string) (*sftp.File, error) {
					return &sftp.File{}, tc.createerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch2)
			patch3, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(f),
				"Close",
				func(c *sftp.File) error {
					return tc.closeerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch3)
			patch4, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(f),
				"Chmod",
				func(c *sftp.File, mode os.FileMode) error {
					return tc.chmoderr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch4)
			patch5, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(f),
				"Write",
				func(c *sftp.File, b []byte) (int, error) {
					return 100, tc.writeerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch5)

			err = WriteRemoteFile(addr, cfg, tc.content, tc.path)
			if err != nil {
				t.Logf("Test Case %s Fail", tc.name)
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}
}

func TestCopyLocalFileToRemoteFile(t *testing.T) {
	t.Log("Start Test CopyLocalFileToRemoteFile")
	errresult := errCase
	_, pwdpath, _, _ := runtime.Caller(0)
	localpath := filepath.Join(filepath.Dir(pwdpath), "testdata/test_rsa")
	cases := []struct {
		name      string
		node      *pluginapi.Node
		local     string
		remote    string
		mkdirerr  error
		clienterr error
		createerr error
		chmoderr  error
		copyerr   error
		returnErr error
	}{
		{"Normal Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		},
		{"Src not exist Case",
			&nodeWithKey,
			"./testdata/fakefile",
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			nil,
			errFileDir,
		},
		{"Dial Err Case",
			&nodeWithWrongPort,
			localpath,
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			nil,
			errConRefuse,
		},
		{"Mkdir Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			errresult,
			nil,
			nil,
			nil,
			nil,
			errresult,
		},
		{"Client Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			nil,
			errresult,
			nil,
			nil,
			nil,
			errresult,
		},
		{"Create Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			nil,
			nil,
			errresult,
			nil,
			nil,
			errresult,
		},
		{"Chmod Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			nil,
			nil,
			nil,
			errresult,
			nil,
			errresult,
		},
		{"Io Copy Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			nil,
			nil,
			nil,
			nil,
			errresult,
			nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c *sftp.Client
			var f *sftp.File
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			mocksftpInterface := fakesftp_mock.NewMockSftpClientInterface(ctrl)
			patchc, err := mpatch.PatchMethod(sftp.NewClient, mocksftpInterface.NewClient)
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patchc)
			mocksftpInterface.EXPECT().NewClient(gomock.Any()).AnyTimes().Return(&sftp.Client{}, tc.clienterr)
			patch, _ := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(c),
				"MkdirAll",
				func(c *sftp.Client, path string) error {
					return tc.mkdirerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch)
			patch2, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(c),
				"Create",
				func(c *sftp.Client, path string) (*sftp.File, error) {
					f = &sftp.File{}
					return f, tc.createerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch2)
			patch3, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(f),
				"Close",
				func(c *sftp.File) error {
					return nil
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch3)
			patch4, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(f),
				"Chmod",
				func(c *sftp.File, mode os.FileMode) error {
					return tc.chmoderr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch4)
			patch5, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(c),
				"Close",
				func(c *sftp.Client) error {
					return nil
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch5)
			patch6, err := mpatch.PatchMethod(io.Copy,
				func(dst io.Writer, src io.Reader) (int64, error) {
					log.Infoln("This should be printed")
					return int64(0), tc.copyerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch6)
			err = CopyLocalFileToRemoteFile(addr, cfg, tc.local, tc.remote)
			if !isExpectedError(err, tc.returnErr) {
				t.Errorf("expected error %v, but function returned error: %v", tc.returnErr, err)
			}
		})
	}
}

func TestCopyRemoteFileToLocalFile(t *testing.T) {
	t.Log("Start Test CopyRemoteFileToLocalFile")
	errresult := errCase
	_, pwdpath, _, _ := runtime.Caller(0)
	localpath := filepath.Join(filepath.Dir(pwdpath), "testdata/testfile")
	wrongpath := filepath.Join(filepath.Dir(pwdpath), "testdata/wrongpath")
	localdir := filepath.Join(filepath.Dir(pwdpath), "testdata")
	cases := []struct {
		name        string
		node        *pluginapi.Node
		remote      string
		local       string
		isDir       bool
		clienterr   error
		staterr     error
		openerr     error
		fileOpenErr error
		osmkdirall  error
		copyErr     error
		returnErr   error
	}{
		{"Normal Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
		},
		{"Dir Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			true,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			errNoFile,
		},

		{"Dial Err Case",
			&nodeWithWrongPort,
			"./test/data/test",
			localpath,
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			errConRefuse,
		},
		{"Client Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			false,
			errresult,
			nil,
			nil,
			nil,
			nil,
			nil,
			errresult,
		},
		{"Stat Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			false,
			nil,
			errresult,
			nil,
			nil,
			nil,
			nil,
			errFileOrDir,
		},
		{"Open Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			false,
			nil,
			nil,
			errresult,
			nil,
			nil,
			nil,
			errresult,
		},
		{"File Open Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			false,
			nil,
			nil,
			nil,
			errresult,
			nil,
			nil,
			errresult,
		},
		{"MkdirAll Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			false,
			nil,
			nil,
			nil,
			nil,
			errresult,
			nil,
			errresult,
		},
		{"Copy Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			false,
			nil,
			nil,
			nil,
			nil,
			nil,
			errresult,
			nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var c *sftp.Client
			var f *sftp.File
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			mocksftpInterface := fakesftp_mock.NewMockSftpClientInterface(ctrl)
			patchc, err := mpatch.PatchMethod(sftp.NewClient, mocksftpInterface.NewClient)
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patchc)
			mocksftpInterface.EXPECT().NewClient(gomock.Any()).AnyTimes().Return(&sftp.Client{}, tc.clienterr)
			patch, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(c),
				"Stat",
				func(c *sftp.Client, path string) (os.FileInfo, error) {
					if tc.staterr != nil {
						return os.Stat(wrongpath)
					}
					if tc.isDir {
						return os.Stat(localdir)
					} else {
						return os.Stat(localpath)
					}
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch)
			patch2, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(c),
				"Open",
				func(c *sftp.Client, path string) (*sftp.File, error) {
					f := &sftp.File{}
					return f, tc.openerr
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch2)
			patch3, err := mpatch.PatchInstanceMethodByName(
				reflect.TypeOf(f),
				"Close",
				func(f *sftp.File) error {
					return nil
				})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch3)
			patch4, err := mpatch.PatchMethod(os.OpenFile, func(name string, flag int, perm os.FileMode) (*os.File, error) {
				return &os.File{}, tc.fileOpenErr
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch4)
			if tc.osmkdirall != nil {
				patch5, err := mpatch.PatchMethod(os.MkdirAll, func(name string, perm os.FileMode) error {
					return tc.osmkdirall
				})
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch5)
			}
			patch6, err := mpatch.PatchMethod(io.Copy, func(dst io.Writer, src io.Reader) (written int64, err error) {
				return 0, tc.copyErr
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch6)
			err = CopyRemoteFileToLocalFile(addr, cfg, tc.remote, tc.local, 0700)
			if !isExpectedError(err, tc.returnErr) {
				t.Errorf("expected error %v, but function returned error: %v", tc.returnErr, err)
			}
		})
	}
}

func TestCopyLocalFileToRemoteRootFileSudoNoPasswd(t *testing.T) {
	t.Log("Start Test CopyLocalFileToRemoteRootFileSudoNoPasswd")
	_, pwdpath, _, _ := runtime.Caller(0)
	localpath := filepath.Join(filepath.Dir(pwdpath), "testdata/test_rsa")
	cases := []struct {
		name           string
		node           *pluginapi.Node
		local          string
		remote         string
		expectedError  error
		funcBeforeTest func(*gomock.Controller)
	}{
		{"Normal Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			nil,
			func(ctrl *gomock.Controller) {
				patchCopyLocalFileToRemoteFile(t, nil)
				patchRunRemoteCMD(t, nil)
			},
		},
		{"Dial Err Case",
			&nodeWithWrongPort,
			localpath,
			"./test/data/test",
			errConRefuse,
			nil,
		},
		{"CopyLocalFileToRemoteFile Err Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			testErr,
			func(ctrl *gomock.Controller) {
				patchCopyLocalFileToRemoteFile(t, testErr)
			},
		},
		{"RunRemoteCMD Err Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			testErr,
			func(ctrl *gomock.Controller) {
				patchCopyLocalFileToRemoteFile(t, nil)
				patchRunRemoteCMD(t, testErr)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest(ctrl)
			}
			err := CopyLocalFileToRemoteRootFileSudoNoPasswd(addr, cfg, tc.local, tc.remote)
			if !isExpectedError(err, tc.expectedError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.expectedError, err)
			}
		})
	}
}

func TestCopyRemoteRootFileToLocalFileSudoNoPasswd(t *testing.T) {
	t.Log("Start Test CopyRemoteRootFileToLocalFileSudoNoPasswd")
	_, pwdpath, _, _ := runtime.Caller(0)
	localpath := filepath.Join(filepath.Dir(pwdpath), "testdata/test_rsa")
	cases := []struct {
		name        string
		node        *pluginapi.Node
		local       string
		remote      string
		copyErr     error
		cmd1Err     error
		cmd2Err     error
		returnError error
	}{
		{"Normal Case",
			&nodeWithKey,
			"./test/data/test",
			localpath,
			nil,
			nil,
			nil,
			nil,
		},
		{"Dial Err Case",
			&nodeWithWrongPort,
			localpath,
			"./test/data/test",
			nil,
			nil,
			nil,
			errConRefuse,
		},
		{"first RunRemoteCMD Err Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			testErr,
			nil,
			nil,
			testErr,
		},
		{"CopyRemoteFileToLocalFile Err Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			nil,
			testErr,
			nil,
			testErr,
		},
		{"second RunRemoteCMD Err Case",
			&nodeWithKey,
			localpath,
			"./test/data/test",
			nil,
			nil,
			testErr,
			testErr,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			patchd, err := mpatch.PatchMethod(CopyRemoteFileToLocalFile, func(addr string, cfg *ssh.ClientConfig, remotePath string, localPath string, perm fs.FileMode) error {
				return tc.copyErr
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patchd)
			var patch *mpatch.Patch
			patch, err = mpatch.PatchMethod(RunRemoteCMD, func(addr string, cfg *ssh.ClientConfig, cmd string) error {
				unpatch(t, patch)
				patch, err = mpatch.PatchMethod(RunRemoteCMD, func(addr string, cfg *ssh.ClientConfig, cmd string) error {
					return tc.cmd2Err
				})
				return tc.cmd1Err
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch)

			err = CopyRemoteRootFileToLocalFileSudoNoPasswd(addr, cfg, tc.remote, tc.local, 0700)
			if !isExpectedError(err, tc.returnError) {
				t.Errorf("expected error %v, but function returned error: %v", tc.returnError, err)
			}

		})
	}
}
func TestContainerdCertificatePathCreateSudoNoPasswd(t *testing.T) {
	t.Log("Start Test ContainerdCertificatePathCreateSudoNoPasswd")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cases := []struct {
		name       string
		node       *pluginapi.Node
		cmd        string
		clienterr  error
		sessionerr error
		cmderr     error
	}{
		{
			"Normal Case",
			&nodeWithKey,
			"echo hi",
			nil,
			nil,
			nil,
		},
		{
			"Client Error",
			&nodeWithWrongPort,
			"echo hi",
			errClnt,
			nil,
			nil,
		},
		{
			"Session Error",
			&nodeWithKey,
			"echo hi",
			nil,
			errSession,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			var c *ssh.Client
			if tc.sessionerr != nil {
				patch, err := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(c),
					"NewSession",
					func(c *ssh.Client) (*ssh.Session, error) {
						return nil, tc.sessionerr
					})
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, patch)
			}
			err := ContainerdCertificatePathCreateSudoNoPasswd(addr, cfg, "", "")
			if err != nil {
				if tc.clienterr != nil || tc.sessionerr != nil || tc.cmderr != nil {
					t.Logf("Test Case %s Pass", tc.name)
				} else {
					t.Logf("Test Case %s Fail", tc.name)
				}
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}

}

func TestServiceRestartSudoNoPasswd(t *testing.T) {
	t.Log("Start Test ServiceRestartSudoNoPasswd")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cases := []struct {
		name       string
		node       *pluginapi.Node
		cmd        string
		clienterr  error
		sessionerr error
		cmderr     error
	}{
		{
			"Normal Case",
			&nodeWithKey,
			"echo hi",
			nil,
			nil,
			nil,
		},
		{
			"Client Error",
			&nodeWithWrongPort,
			"echo hi",
			errClnt,
			nil,
			nil,
		},
		{
			"Session Error",
			&nodeWithKey,
			"echo hi",
			nil,
			errSession,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			var c *ssh.Client
			if tc.sessionerr != nil {
				guard, _ := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(c),
					"NewSession",
					func(c *ssh.Client) (*ssh.Session, error) {
						return nil, tc.sessionerr
					})
				defer unpatch(t, guard)
			}
			err := ServiceRestartSudoNoPasswd(addr, cfg, "")
			if err != nil {
				if tc.clienterr != nil || tc.sessionerr != nil || tc.cmderr != nil {
					t.Logf("Test Case %s Pass", tc.name)
				} else {
					t.Logf("Test Case %s Fail", tc.name)
				}
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}

}

func TestRemoteFileExists(t *testing.T) {
	t.Log("Start Test RemoteFileExists")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cases := []struct {
		name       string
		node       *pluginapi.Node
		directory  string
		clienterr  error
		sessionerr error
		cmderr     error
	}{
		{
			"Client Error",
			&nodeWithWrongPort,
			"/tmp",
			errClnt,
			nil,
			nil,
		},
		{
			"Session Error",
			&nodeWithKey,
			"/tmp",
			nil,
			errSession,
			nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("Run Test Case %s", tc.name)
			cfg, _ := GenSSHConfig(tc.node)
			addr := fmt.Sprintf("%s:%d", tc.node.IP, tc.node.SSHPort)
			var c *ssh.Client
			if tc.sessionerr != nil {
				guard, _ := mpatch.PatchInstanceMethodByName(
					reflect.TypeOf(c),
					"NewSession",
					func(c *ssh.Client) (*ssh.Session, error) {
						return nil, tc.sessionerr
					})
				defer unpatch(t, guard)
			}
			_, err := RemoteFileExists(addr, cfg, "")
			if err != nil {
				if tc.clienterr != nil || tc.sessionerr != nil || tc.cmderr != nil {
					t.Logf("Test Case %s Pass", tc.name)
				} else {
					t.Logf("Test Case %s Fail", tc.name)
				}
			} else {
				t.Logf("Test Case %s Pass", tc.name)
			}
		})
	}

}

func runSSHServer() {
	addr := fmt.Sprintf("%s:%d", sshHost, sshPort)
	sshd.Handle(func(s sshd.Session) {
		_, err := io.WriteString(s, "Hello world\n")
		if err != nil {
			log.Errorln(err)
		}
	})
	log.Fatal(sshd.ListenAndServe(addr, nil))
}

func TestMain(m *testing.M) {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath := filepath.Join(filepath.Dir(pwdpath), "testdata")
	keyfile := filepath.Join(testdatapath, "test_rsa")
	SSHKey, err := ioutil.ReadFile(keyfile)
	keystring := string(SSHKey)
	if err != nil {
		log.Error(err)
	}
	nodeWithKey.SSHKey = keystring
	go runSSHServer()
	os.Exit(m.Run())
}
