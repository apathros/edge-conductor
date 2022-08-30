/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package capiprovisionbinarydownload

import (
	"errors"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	capiutils "github.com/intel/edge-conductor/pkg/eputils/capiutils"
	"github.com/intel/edge-conductor/pkg/eputils/docker"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

var (
	errEmpty = errors.New("")
)

/**
 * Test function PluginMain
 **/
func Test_PluginMain(t *testing.T) {
	var cases = []struct {
		name           string
		expectError    error
		in             eputils.SchemaMapData
		outp           *eputils.SchemaMapData
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:        "Get Capi Template Failed",
			expectError: errEmpty,
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Cluster: &pluginapi.KitconfigCluster{
							Config: "",
						},
						Parameters: &pluginapi.KitconfigParameters{Extensions: []string{"capi-byoh"}},
					},
				}

				data[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
				return data
			}(),
			outp: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchCreateFolderIfNotExist(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile2(t, false)
				patch4 := patchGetCapiSetting(t, false)
				patch5 := patchGetCapiTemplate(t, true)
				patch7 := patchMkdirAll(t, false)
				patch8 := patchLaunchIpaDownload(t, false)
				patch9 := patchCopyIronicProvisionImage(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch7, patch8, patch9}
			},
		},
		{
			name:        "No provider",
			expectError: eputils.GetError("errProvider"),
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Parameters: &pluginapi.KitconfigParameters{
							Extensions: []string{},
						},
					},
				}

				data[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
				return data
			}(),
			outp: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchCreateFolderIfNotExist(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile2(t, false)
				patch4 := patchGetCapiSetting(t, false)
				patch5 := patchGetCapiTemplate(t, false)
				patch7 := patchMkdirAll(t, false)
				patch8 := patchLaunchIpaDownload(t, false)
				patch9 := patchCopyIronicProvisionImage(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch7, patch8, patch9}
			},
		},
		{
			name:        "Launch IPA Download Failed",
			expectError: errEmpty,
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Cluster: &pluginapi.KitconfigCluster{
							Config: "",
						},
						Parameters: &pluginapi.KitconfigParameters{Extensions: []string{"capi-metal3"}},
					},
				}

				data[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
				return data
			}(),
			outp: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchCreateFolderIfNotExist(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile2(t, false)
				patch4 := patchGetCapiSetting(t, false)
				patch5 := patchGetCapiTemplate(t, false)
				patch7 := patchMkdirAll(t, false)
				patch8 := patchLaunchIpaDownload(t, true)
				patch9 := patchCopyIronicProvisionImage(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch7, patch8, patch9}
			},
		},
		{
			name:        "Copy Ironic OS Image Failed",
			expectError: errEmpty,
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Cluster: &pluginapi.KitconfigCluster{
							Config: "",
						},
						Parameters: &pluginapi.KitconfigParameters{Extensions: []string{"capi-metal3"}},
					},
				}

				data[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
				return data
			}(),
			outp: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchCreateFolderIfNotExist(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile2(t, false)
				patch4 := patchGetCapiSetting(t, false)
				patch5 := patchGetCapiTemplate(t, false)
				patch7 := patchMkdirAll(t, false)
				patch8 := patchLaunchIpaDownload(t, false)
				patch9 := patchCopyIronicProvisionImage(t, true)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch7, patch8, patch9}
			},
		},
		{
			name:        "Load Schema Struct Failed",
			expectError: errEmpty,
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Cluster: &pluginapi.KitconfigCluster{
							Config: "",
						},
						Parameters: &pluginapi.KitconfigParameters{Extensions: []string{"capi-metal3"}},
					},
				}

				data[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
				return data
			}(),
			outp: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchCreateFolderIfNotExist(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile2(t, true)
				patch4 := patchGetCapiSetting(t, false)
				patch5 := patchGetCapiTemplate(t, false)
				patch7 := patchMkdirAll(t, false)
				patch8 := patchLaunchIpaDownload(t, false)
				patch9 := patchCopyIronicProvisionImage(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch7, patch8, patch9}
			},
		},
		{
			name:        "success, byoh provider",
			expectError: nil,
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Cluster: &pluginapi.KitconfigCluster{
							Config: "",
						},
						Parameters: &pluginapi.KitconfigParameters{Extensions: []string{"capi-byoh"}},
					},
				}

				data[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
				return data
			}(),
			outp: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchCreateFolderIfNotExist(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile2(t, false)
				patch4 := patchGetCapiSetting(t, false)
				patch5 := patchGetCapiTemplate(t, false)
				patch7 := patchMkdirAll(t, false)
				patch8 := patchLaunchIpaDownload(t, false)
				patch9 := patchCopyIronicProvisionImage(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch7, patch8, patch9}
			},
		},
		{
			name:        "success, metal3 provider",
			expectError: nil,
			in: func() eputils.SchemaMapData {
				data := eputils.NewSchemaMapData()
				data[__name("ep-params")] = &pluginapi.EpParams{
					Kitconfig: &pluginapi.Kitconfig{
						Cluster: &pluginapi.KitconfigCluster{
							Config: "",
						},
						Parameters: &pluginapi.KitconfigParameters{Extensions: []string{"capi-metal3"}},
					},
				}

				data[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
				return data
			}(),
			outp: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchCreateFolderIfNotExist(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile2(t, false)
				patch4 := patchGetCapiSetting(t, false)
				patch5 := patchGetCapiTemplate(t, false)
				patch7 := patchMkdirAll(t, false)
				patch8 := patchLaunchIpaDownload(t, false)
				patch9 := patchCopyIronicProvisionImage(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch7, patch8, patch9}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := PluginMain(tc.in, tc.outp)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func patchCopyIronicProvisionImage(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(copyIronicOsPrivisionImage, func(ep_params *pluginapi.EpParams, ironicHttpdFolder string, capiSetting *pluginapi.CapiSetting) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchLaunchIpaDownload(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(launchIpaDownload, func(workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

//nolint:unparam
func patchMkdirAll(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(os.MkdirAll, func(path string, perm os.FileMode) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchGetCapiTemplate(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(capiutils.GetCapiTemplate, func(epparams *pluginapi.EpParams, setting pluginapi.CapiSetting, cp *capiutils.CapiTemplate) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

//nolint:unparam
func patchGetCapiSetting(t *testing.T, false bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(capiutils.GetCapiSetting, func(epparams *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, clusterConfig *pluginapi.CapiClusterConfig, setting *pluginapi.CapiSetting) error {
		return nil
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

//nolint:unparam
func patchCreateFolderIfNotExist(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.CreateFolderIfNotExist, func(path string) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func isExpectedError(returnErr error, wantError error) bool {
	if !errors.Is(returnErr, wantError) &&
		(returnErr == nil || wantError == nil || !strings.Contains(returnErr.Error(), wantError.Error())) {
		return false
	}
	return true
}

/**
 * Test function copyIronicOsPrivisionImage
 **/
func Test_copyIronicOsPrivisionImage(t *testing.T) {
	var cases = []struct {
		name              string
		expectError       error
		epparam           *pluginapi.EpParams
		ironicHttpdFolder string
		capiSetting       *pluginapi.CapiSetting
		funcBeforeTest    func() []*mpatch.Patch
	}{
		{
			name:              "File does not exist",
			expectError:       nil,
			epparam:           &pluginapi.EpParams{},
			ironicHttpdFolder: "provisionOsImage",
			capiSetting: &pluginapi.CapiSetting{
				IronicConfig: &pluginapi.CapiSettingIronicConfig{
					IronicOsImage: "Ubuntu",
				},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathBase(t)
				patch3 := patchFileExists(t, false)
				patch4 := patchCopyFile(t, false)
				patch5 := patchGenFileSHA256(t, false)
				patch6 := patchWriteStringToFile(t, false)
				return []*mpatch.Patch{patch1, patch3, patch4, patch5, patch6}
			},
		},
		{
			name:              "success",
			expectError:       nil,
			epparam:           &pluginapi.EpParams{},
			ironicHttpdFolder: "",
			capiSetting: &pluginapi.CapiSetting{
				IronicConfig: &pluginapi.CapiSettingIronicConfig{
					IronicOsImage: "Ubuntu",
				},
			},
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathBase(t)
				patch2 := patchFilePathJoin(t)
				patch3 := patchFileExists(t, false)
				patch4 := patchCopyFile(t, false)
				patch5 := patchGenFileSHA256(t, false)
				patch6 := patchWriteStringToFile(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4, patch5, patch6}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.funcBeforeTest != nil {
				pList := tc.funcBeforeTest()
				defer unpatchAll(t, pList)
			}

			err := copyIronicOsPrivisionImage(tc.epparam, tc.ironicHttpdFolder, tc.capiSetting)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

//nolint:unparam
func patchFileExists(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.FileExists, func(filename string) bool {
		if fail {
			return false
		} else {
			return filename != "provisionOsImage"
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchWriteStringToFile(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.WriteStringToFile, func(content, filename string) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchGenFileSHA256(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.GenFileSHA256, func(filename string) (string, error) {
		if fail {
			return "", errEmpty
		} else {
			return "", nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchCopyFile(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.CopyFile, func(dstName, srcName string) (written int64, err error) {
		if fail {
			return -1, errEmpty
		} else {
			return 1, nil
		}
	})
	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchFilePathBase(t *testing.T) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(filepath.Base, func(path string) string {
		return ""

	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

/**
 * Test function launchIpaDownload
 **/
func Test_launchIpaDownload(t *testing.T) {
	var cases = []struct {
		name           string
		epparam        *pluginapi.EpParams
		workfolder     string
		clusterconfig  *pluginapi.CapiClusterConfig
		tmpl           *capiutils.CapiTemplate
		expectError    error
		funcBeforeTest func() []*mpatch.Patch
	}{
		{
			name:       "success",
			epparam:    &pluginapi.EpParams{},
			workfolder: "",
			clusterconfig: &pluginapi.CapiClusterConfig{
				BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{},
			},
			tmpl: &capiutils.CapiTemplate{
				EpParams:    pluginapi.EpParams{},
				CapiSetting: pluginapi.CapiSetting{},
			},
			expectError: nil,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchTmplFileRendering(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile1(t, false)
				patch4 := patchDockerRun(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4}
			},
		},
		{
			name:       "Template File Rendering Failed",
			epparam:    &pluginapi.EpParams{},
			workfolder: "",
			clusterconfig: &pluginapi.CapiClusterConfig{
				BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{},
			},
			tmpl: &capiutils.CapiTemplate{
				EpParams:    pluginapi.EpParams{},
				CapiSetting: pluginapi.CapiSetting{},
			},
			expectError: errEmpty,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchTmplFileRendering(t, true)
				patch3 := patchLoadSchemaStructFromYamlFile1(t, false)
				patch4 := patchDockerRun(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4}
			},
		},
		{
			name:       "Load Schema Struct from yaml failed",
			epparam:    &pluginapi.EpParams{},
			workfolder: "",
			clusterconfig: &pluginapi.CapiClusterConfig{
				BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{},
			},
			tmpl: &capiutils.CapiTemplate{
				EpParams:    pluginapi.EpParams{},
				CapiSetting: pluginapi.CapiSetting{},
			},
			expectError: errEmpty,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchTmplFileRendering(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile1(t, true)
				patch4 := patchDockerRun(t, false)
				return []*mpatch.Patch{patch1, patch2, patch3, patch4}
			},
		},
		{
			name:       "Docker Run failed",
			epparam:    &pluginapi.EpParams{},
			workfolder: "",
			clusterconfig: &pluginapi.CapiClusterConfig{
				BaremetelOperator: &pluginapi.CapiClusterConfigBaremetelOperator{},
			},
			tmpl: &capiutils.CapiTemplate{
				EpParams:    pluginapi.EpParams{},
				CapiSetting: pluginapi.CapiSetting{},
			},
			expectError: errEmpty,
			funcBeforeTest: func() []*mpatch.Patch {
				patch1 := patchFilePathJoin(t)
				patch2 := patchTmplFileRendering(t, false)
				patch3 := patchLoadSchemaStructFromYamlFile1(t, false)
				patch4 := patchDockerRun(t, true)
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

			err := launchIpaDownload(tc.workfolder, tc.clusterconfig, tc.tmpl)

			if !isExpectedError(err, tc.expectError) {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func patchDockerRun(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(docker.DockerRun, func(in_config *pluginapi.ContainersItems0) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchLoadSchemaStructFromYamlFile1(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
		if fail {

			return errEmpty
		} else {
			v.(*pluginapi.Containers).Containers = []*pluginapi.ContainersItems0{
				{
					Name: "ipa-downloader",
				},
			}
			return nil
		}

	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchLoadSchemaStructFromYamlFile2(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(eputils.LoadSchemaStructFromYamlFile, func(v eputils.SchemaStruct, file string) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}

	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchTmplFileRendering(t *testing.T, fail bool) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(capiutils.TmplFileRendering, func(tmpl *capiutils.CapiTemplate, workFolder, url, dstFile string) error {
		if fail {
			return errEmpty
		} else {
			return nil
		}
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func patchFilePathJoin(t *testing.T) *mpatch.Patch {
	patch, patchErr := mpatch.PatchMethod(filepath.Join, func(elem ...string) string {
		return ""
	})

	if patchErr != nil {
		t.Errorf("patch error: %v", patchErr)
	}

	return patch
}

func unpatchAll(t *testing.T, pList []*mpatch.Patch) {
	for _, p := range pList {
		if p != nil {
			if err := p.Unpatch(); err != nil {
				t.Errorf("unpatch error: %v", err)
			}
		}
	}
}

func TestInitStructFunc(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "ep-params",
			input: __name("ep-params"),
		},
		{
			name:  "cluster-manifest",
			input: __name("cluster-manifest"),
		},
	}

	// Optional: add setup for the test series
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eputils.SchemaStructNew(tc.input)

		})
	}
}
