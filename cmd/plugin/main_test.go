/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package main

import (
	epapp "ep/cmd/ep/app"
	epapiplugins "ep/pkg/api/plugins"
	plugin "ep/pkg/plugin"
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errTest = fmt.Errorf("test error")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			unpatch(t, p)
		}
	}
}

func patchOsExit(t *testing.T) *mpatch.Patch {
	exitGuard, patchErr := mpatch.PatchMethod(os.Exit, func(code int) {
		panic(code)
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return exitGuard
}

func patchFlagArgs(t *testing.T, args []string) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(flag.Args, func() []string {
		return args
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchInitConnect(t *testing.T, err error) *mpatch.Patch {
	epparams := &epapiplugins.EpParams{}
	pInit := plugin.New("__init__", epparams, nil)
	patch, patchErr := mpatch.PatchInstanceMethodByName(reflect.TypeOf(pInit), "Connect", func(p *plugin.Plugin, addr string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchEpUtilsInit(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(epapp.EpUtilsInit, func(epParams *epapiplugins.EpParams) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchEnablePluginRemoteLog(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(plugin.EnablePluginRemoteLog, func(name string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchStartPlugin(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(plugin.StartPlugin, func(name string, errch chan error) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func patchWaitPluginFinished(t *testing.T, err error) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(plugin.WaitPluginFinished, func(name string) error {
		return err
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
		return nil
	}
	return patch
}

func TestMainFunction(t *testing.T) {
	cases := []struct {
		name           string
		funcBeforeTest func() []*mpatch.Patch
		wantExitCode   interface{}
	}{
		{
			name: "Empty Args",
			funcBeforeTest: func() []*mpatch.Patch {
				pos := patchOsExit(t)
				pargs := patchFlagArgs(t, []string{})
				return []*mpatch.Patch{pos, pargs}
			},
			wantExitCode: 1,
		},
		{
			name: "Init Connect Failure",
			funcBeforeTest: func() []*mpatch.Patch {
				pos := patchOsExit(t)
				pargs := patchFlagArgs(t, []string{"fakeaddr"})
				pconnect := patchInitConnect(t, errTest)
				return []*mpatch.Patch{pos, pargs, pconnect}
			},
			wantExitCode: 1,
		},
		{
			name: "EpUtilsInit Failure",
			funcBeforeTest: func() []*mpatch.Patch {
				pos := patchOsExit(t)
				pargs := patchFlagArgs(t, []string{"fakeaddr"})
				pconnect := patchInitConnect(t, nil)
				puinit := patchEpUtilsInit(t, errTest)
				return []*mpatch.Patch{pos, pargs, pconnect, puinit}
			},
			wantExitCode: 1,
		},
		{
			name: "EnablePluginRemoteLog Failure",
			funcBeforeTest: func() []*mpatch.Patch {
				pos := patchOsExit(t)
				pargs := patchFlagArgs(t, []string{"fakeaddr"})
				pconnect := patchInitConnect(t, nil)
				puinit := patchEpUtilsInit(t, nil)
				prlog := patchEnablePluginRemoteLog(t, errTest)
				return []*mpatch.Patch{pos, pargs, pconnect, puinit, prlog}
			},
			wantExitCode: 1,
		},
		{
			name: "StartPlugin Failure",
			funcBeforeTest: func() []*mpatch.Patch {
				pos := patchOsExit(t)
				pargs := patchFlagArgs(t, []string{"fakeaddr"})
				pconnect := patchInitConnect(t, nil)
				puinit := patchEpUtilsInit(t, nil)
				prlog := patchEnablePluginRemoteLog(t, nil)
				pstart := patchStartPlugin(t, errTest)
				return []*mpatch.Patch{pos, pargs, pconnect, puinit, prlog, pstart}
			},
			wantExitCode: 1,
		},
		{
			name: "WaitPluginFinished Failure",
			funcBeforeTest: func() []*mpatch.Patch {
				pos := patchOsExit(t)
				pargs := patchFlagArgs(t, []string{"fakeaddr"})
				pconnect := patchInitConnect(t, nil)
				puinit := patchEpUtilsInit(t, nil)
				prlog := patchEnablePluginRemoteLog(t, nil)
				pstart := patchStartPlugin(t, nil)
				pwait := patchWaitPluginFinished(t, errTest)
				return []*mpatch.Patch{pos, pargs, pconnect, puinit, prlog, pstart, pwait}
			},
			wantExitCode: 1,
		},
		{
			name: "Success",
			funcBeforeTest: func() []*mpatch.Patch {
				pos := patchOsExit(t)
				pargs := patchFlagArgs(t, []string{"fakeaddr"})
				pconnect := patchInitConnect(t, nil)
				puinit := patchEpUtilsInit(t, nil)
				prlog := patchEnablePluginRemoteLog(t, nil)
				pstart := patchStartPlugin(t, nil)
				pwait := patchWaitPluginFinished(t, nil)
				return []*mpatch.Patch{pos, pargs, pconnect, puinit, prlog, pstart, pwait}
			},
			wantExitCode: nil,
		},
	}

	for _, testCase := range cases {
		t.Logf("TestMainFunction case %s start", testCase.name)
		func() {
			if testCase.funcBeforeTest != nil {
				pList := testCase.funcBeforeTest()
				defer unpatchAll(t, pList)
			}
			defer func() {
				exitCode := recover()
				if !reflect.DeepEqual(exitCode, testCase.wantExitCode) {
					t.Errorf("Unexpected exit code: %v", exitCode)
				}
			}()
			main()
		}()
		t.Logf("TestMainFunction case %s end", testCase.name)
	}
	t.Log("Done")
}
