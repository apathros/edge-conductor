/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package executor

import (
	"context"
	"errors"
	"io"
	"net"
	"reflect"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
	"golang.org/x/crypto/ssh"
)

var (
	errSSHTestEmpty = errors.New("")
	errClose        = errors.New("close.error")
)

func TestConnect(t *testing.T) {
	var cases = []struct {
		name           string
		sshclient      *sshClient
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name: "success",
			sshclient: &sshClient{
				client: &ssh.Client{},
			},
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchParsePrivateKey(t, false)
				patch2 := patchDial(t, false)
				return []*mpatch.Patch{patch1, patch2}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := tc.sshclient.Connect()
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

func patchclientclose(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	var sshclient = &sshClient{
		client: &ssh.Client{}}
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(sshclient.client), "Close", func(*ssh.Client) error {
		if ok {
			return nil
		} else {
			return errClose
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

type fakeConn struct{}

func (*fakeConn) SendRequest(name string, wantReply bool, payload []byte) (bool, []byte, error) {
	return false, nil, nil
}
func (*fakeConn) OpenChannel(name string, data []byte) (ssh.Channel, <-chan *ssh.Request, error) {
	return nil, nil, nil
}
func (*fakeConn) Close() error          { return nil }
func (*fakeConn) Wait() error           { return nil }
func (*fakeConn) User() string          { return "nil" }
func (*fakeConn) SessionID() []byte     { return nil }
func (*fakeConn) ClientVersion() []byte { return nil }
func (*fakeConn) ServerVersion() []byte { return nil }
func (*fakeConn) RemoteAddr() net.Addr  { return nil }
func (*fakeConn) LocalAddr() net.Addr   { return nil }

func TestDisconnect(t *testing.T) {
	var p1 *mpatch.Patch
	var cases = []struct {
		name           string
		sshclient      *sshClient
		expectError    bool
		funcBeforeTest func()
		teardown       func()
	}{
		{
			name: "Disconnect success",
			sshclient: &sshClient{
				client: &ssh.Client{
					Conn: &fakeConn{},
				},
			},
			expectError: false,
			funcBeforeTest: func() {
				p1 = patchclientclose(t, true)
			},
			teardown: func() {
				unpatch(t, p1)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				tc.funcBeforeTest()
			}

			err := tc.sshclient.Disconnect()
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}

			if tc.teardown != nil {
				tc.teardown()
			}
		})
	}
}

func patchDial(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(ssh.Dial, func(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error) {
		if fail {
			return nil, errSSHTestEmpty
		} else {
			return nil, nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchParsePrivateKey(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(ssh.ParsePrivateKey, func(pemBytes []byte) (ssh.Signer, error) {
		if fail {
			return nil, errSSHTestEmpty
		} else {
			return nil, nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestCmdWithAttachIO(t *testing.T) {
	var cases = []struct {
		name           string
		sshclient      *sshClient
		ctx            context.Context
		cmd            []string
		stdin          io.Reader
		stdout         io.Writer
		stderr         io.Writer
		tty            bool
		expectError    bool
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name: "has stdin/stdout/stderr",
			sshclient: &sshClient{
				client: &ssh.Client{},
			},
			ctx:         context.TODO(),
			stdin:       &MockReader{},
			stdout:      &MockWriter{},
			stderr:      &MockWriter{},
			tty:         true,
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchNewSession(t, false)
				patch2 := patchSessionClose(t, false)
				patch3 := patchRequestPty(t, false)
				patch4 := patchSessionRun(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4}
			},
		},
		{
			name: "request pty fail",
			sshclient: &sshClient{
				client: &ssh.Client{},
			},
			ctx:         context.TODO(),
			tty:         true,
			expectError: true,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchNewSession(t, false)
				patch2 := patchSessionClose(t, false)
				patch3 := patchRequestPty(t, true)
				patch4 := patchSessionRun(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4}
			},
		},
		{
			name: "request pty success",
			sshclient: &sshClient{
				client: &ssh.Client{},
			},
			ctx:         context.TODO(),
			tty:         true,
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchNewSession(t, false)
				patch2 := patchSessionClose(t, false)
				patch3 := patchRequestPty(t, false)
				patch4 := patchSessionRun(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4}
			},
		},
		{
			name: "success",
			sshclient: &sshClient{
				client: &ssh.Client{},
			},
			ctx:         context.TODO(),
			tty:         false,
			expectError: false,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchNewSession(t, false)
				patch2 := patchSessionClose(t, false)
				patch3 := patchRequestPty(t, false)
				patch4 := patchSessionRun(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := tc.sshclient.CmdWithAttachIO(tc.ctx, tc.cmd, tc.stdin, tc.stdout, tc.stderr, tc.tty)
			if (err != nil && !tc.expectError) ||
				(err == nil && tc.expectError) {
				t.Logf("Test case %s failed.", tc.name)
				t.Error(err)
			} else {
				t.Log("Done")
			}
		})
	}
}

//nolint:unparam
func patchSessionRun(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&ssh.Session{}), "Run", func(s *ssh.Session, cmd string) error {
		if fail {
			return errSSHTestEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchRequestPty(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&ssh.Session{}), "RequestPty", func(s *ssh.Session, term string, h, w int, termmodes ssh.TerminalModes) error {
		if fail {
			return errSSHTestEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

//nolint:unparam
func patchSessionClose(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&ssh.Session{}), "Close", func(s *ssh.Session) error {
		if fail {
			return errSSHTestEmpty
		} else {
			return nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

//nolint:unparam
func patchNewSession(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(&ssh.Client{}), "NewSession", func(s *ssh.Client) (*ssh.Session, error) {
		if fail {
			return nil, errSSHTestEmpty
		} else {
			return &ssh.Session{}, nil
		}

	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}
