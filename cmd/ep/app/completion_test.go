/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	"errors"
	"io"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
	"github.com/undefinedlabs/go-mpatch"
)

var (
	errGenBash = errors.New("genbashcompletion.error")
	errGenZsh  = errors.New("genzshcompletion.error")
	errGenFish = errors.New("genfishcompletion.error")
	errGenPwsh = errors.New("genpwshcompletion.error")
)

//nolint:unparam
func patchcmdroot(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&cobra.Command{}), "Root", func(c *cobra.Command) *cobra.Command {
		if ok {
			return &cobra.Command{}
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchgenbashcompleton(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&cobra.Command{}), "GenBashCompletion", func(c *cobra.Command, w io.Writer) error {
		if ok {
			return nil
		} else {
			return errGenBash
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchgenzshcompletion(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&cobra.Command{}), "GenZshCompletion", func(c *cobra.Command, w io.Writer) error {
		if ok {
			return nil

		} else {
			return errGenZsh
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchgenfishcompletion(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&cobra.Command{}), "GenFishCompletion", func(c *cobra.Command, w io.Writer, includeDesc bool) error {
		if ok {
			return nil

		} else {
			return errGenFish
		}

	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchgenpwshcompletion(t *testing.T, ok bool) *mpatch.Patch {
	var patch *mpatch.Patch
	var patchErr error
	patch, patchErr = mpatch.PatchInstanceMethodByName(reflect.TypeOf(&cobra.Command{}), "GenPowerShellCompletion", func(c *cobra.Command, w io.Writer) error {
		if ok {
			return nil
		} else {
			return errGenPwsh
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func TestCompletionCMD(t *testing.T) {
	var p1 *mpatch.Patch
	var p2 *mpatch.Patch
	cases := []struct {
		name       string
		shell_type string
		beforetest func()
		teardown   func()
	}{
		{
			name:       "User is running bash",
			shell_type: "bash",
			beforetest: func() {
				p1 = patchcmdroot(t, true)
				p2 = patchgenbashcompleton(t, false)
			},
			teardown: func() {
				unpatch(t, p1)
				unpatch(t, p2)
			},
		},
		{
			name:       "User is running zsh",
			shell_type: "zsh",
			beforetest: func() {
				p1 = patchcmdroot(t, true)
				p2 = patchgenzshcompletion(t, false)
			},
			teardown: func() {
				unpatch(t, p1)
				unpatch(t, p2)
			},
		},
		{
			name:       "User is running fish",
			shell_type: "fish",
			beforetest: func() {
				p1 = patchcmdroot(t, true)
				p2 = patchgenfishcompletion(t, false)
			},
			teardown: func() {
				unpatch(t, p1)
				unpatch(t, p2)
			},
		},
		{
			name:       "User is running powershell",
			shell_type: "powershell",
			beforetest: func() {
				p1 = patchcmdroot(t, true)
				p2 = patchgenpwshcompletion(t, false)
			},
			teardown: func() {
				unpatch(t, p1)
				unpatch(t, p2)
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			for _, tc := range cases {
				if tc.beforetest != nil {
					tc.beforetest()
				}

				completionCmd.Run(&cobra.Command{}, []string{tc.shell_type})

				if tc.teardown != nil {
					tc.teardown()
				}
			}

		})
	}
}
