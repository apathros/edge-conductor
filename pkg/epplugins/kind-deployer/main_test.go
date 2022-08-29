/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package kinddeployer

import (
	"encoding/json"
	"ep/pkg/eputils"
	"ep/pkg/eputils/repoutils"
	"errors"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	errEmpty = errors.New("")
)

var _ = Describe("Main", func() {
	patchService := NewPatchService()

	patchService.CreateNewPatch("os_mkdirall_fail", os.MkdirAll, func(path string, perm fs.FileMode) error {
		return errEmpty
	}).CreateNewPatch("os_mkdirall_ok", os.MkdirAll, func(path string, perm fs.FileMode) error {
		return nil
	}).CreateNewPatch("os_chmod_fail", os.Chmod, func(name string, mode fs.FileMode) error {
		return errEmpty
	}).CreateNewPatch("os_chmod_ok", os.Chmod, func(name string, mode fs.FileMode) error {
		return nil
	}).CreateNewPatch("eputil_runcmdex_ok", eputils.RunCMDEx, func(cmd *exec.Cmd, useOsOut bool) (string, error) {
		return "success", nil
	}).CreateNewPatch("ioutil_readfile_ok", ioutil.ReadFile, func(filename string) ([]byte, error) {
		return []byte("Success"), nil
	}).CreateNewPatch("repoutils_pullfilefromrepo_fail", repoutils.PullFileFromRepo, func(filepath string, targeturl string) error {
		return errEmpty
	}).CreateNewPatch("os.Stat_ok", os.Stat, func(name string) (fs.FileInfo, error) {
		return nil, nil
	}).CreateNewPatch("os.Stat_fail", os.Stat, func(name string) (fs.FileInfo, error) {
		return nil, errEmpty
	}).CreateNewPatch("os.IsNotExist_y", os.IsNotExist, func(err error) bool {
		return true
	}).CreateNewPatch("os.IsNotExist_n", os.IsNotExist, func(err error) bool {
		return false
	}).CreateNewPatch("os_mkdirall_1", os.MkdirAll, func(path string, perm fs.FileMode) error {
		if perm == 0700 {
			return nil
		} else if perm == 0777 {
			return errEmpty
		} else {
			return nil
		}
	}).CreateNewPatch("ioutil.WriteFile_ok", ioutil.WriteFile, func(filename string, data []byte, perm fs.FileMode) error {
		return nil
	}).CreateNewPatch("ioutil.WriteFile_fail", ioutil.WriteFile, func(filename string, data []byte, perm fs.FileMode) error {
		return errEmpty
	}).CreateNewPatch("eputil_runcmdex_fail", eputils.RunCMDEx, func(cmd *exec.Cmd, useOsOut bool) (string, error) {
		return "", errEmpty
	}).CreateNewPatch("eputils.RemoveFile_ok", eputils.RemoveFile, func(name string) error {
		return nil
	}).CreateNewPatch("eputils.RemoveFile_fail", eputils.RemoveFile, func(name string) error {
		return errEmpty
	}).CreateNewPatch("ioutil_readfile_fail", ioutil.ReadFile, func(filename string) ([]byte, error) {
		return nil, errEmpty
	}).CreateNewPatch("os.mkdir_ok", os.Mkdir, func(name string, perm fs.FileMode) error {
		return nil
	})

	//region Test Assertions (Ω)
	var (
		AssertInputCheckFail = func(input map[string][]byte) func() {
			return func() {
				data := generateInput(input)
				Ω(data).Should(BeNil())
			}
		}

		AssertDeployFail = func(ctx *PatchRequestContext, input map[string][]byte, output eputils.SchemaMapData) func() {
			return func() {
				ctx.printDebugSummary()
				patchService.PatchRequested(ctx)
				err := PluginMain(generateInput(input), &output)
				Ω(err).ShouldNot(BeNil())
				patchService.UnpatchRequested(ctx)
			}
		}

		AssertDeploySuccess = func(ctx *PatchRequestContext, input map[string][]byte, output eputils.SchemaMapData) func() {
			return func() {
				ctx.printDebugSummary()
				patchService.PatchRequested(ctx)
				err := PluginMain(generateInput(input), &output)
				Ω(err).Should(BeNil())
				patchService.UnpatchRequested(ctx)
			}
		}
	)
	//endregion

	//region Test Body
	Context("Test Body", func() {
		const (
			testdata_input_filepath  = "testdata/input.json"
			testdata_output_filepath = "testdata/output.json"
		)
		var (
			input_template_good  map[string][]byte
			output_template_good map[string][]byte
			output               eputils.SchemaMapData
		)
		var (
			unmarshaled_map map[string]interface{}
		)
		if b, err := ioutil.ReadFile(testdata_input_filepath); err != nil {
			Fail("[Fatal]: Cannot read testdata from `" + testdata_input_filepath + "`.")
		} else {
			if err := json.Unmarshal(b, &unmarshaled_map); err != nil {
				Fail("[Fatal]: Cannot parse testdata from `" + testdata_input_filepath + "`.")
			} else {
				input_template_good = make(map[string][]byte)

				for k, v := range unmarshaled_map {
					if b, err := json.Marshal(v); err != nil {
						Fail("[Fatal]: Cannot remarshal testdata from `" + testdata_input_filepath + "`, but it was parsed successfully. ")
					} else {
						input_template_good[k] = b
					}

				}
			}
		}

		if b, err := ioutil.ReadFile(testdata_output_filepath); err != nil {
			Fail("[Fatal]: Cannot read testdata from `" + testdata_output_filepath + "`.")
		} else {
			if err := json.Unmarshal(b, &unmarshaled_map); err != nil {
				Fail("[Fatal]: Cannot parse testdata from `" + testdata_output_filepath + "`.")
			} else {
				output_template_good = make(map[string][]byte)

				for k, v := range unmarshaled_map {
					if b, err := json.Marshal(v); err != nil {
						Fail("[Fatal]: Cannot remarshal testdata from `" + testdata_output_filepath + "`, but it was parsed successfully. ")
					} else {
						output_template_good[k] = b
					}

				}

				output = generateOutput(output_template_good)
			}
		}

		genGoodInput := func() map[string][]byte {
			input_good := make(map[string][]byte)
			for key, value := range input_template_good {
				input_good[key] = value
			}

			return input_good
		}

		When("Input is NOT valid", func() {
			input_bad_files := make(map[string][]byte)
			for key, value := range input_template_good {
				input_bad_files[key] = value
			}
			input_bad_files["files"] = []uint8{1, 2, 3}
			It("Should fail", AssertInputCheckFail(input_bad_files))

			input_bad_ep_params := make(map[string][]byte)
			for key, value := range input_template_good {
				input_bad_ep_params[key] = value
			}
			input_bad_ep_params["ep-params"] = []uint8{1, 2, 3}
			It("Should fail", AssertInputCheckFail(input_bad_ep_params))

			input_bad_kind_config := make(map[string][]byte)
			for key, value := range input_template_good {
				input_bad_kind_config[key] = value
			}
			input_bad_kind_config["kind-config"] = []uint8{1, 2, 3}
			It("Should fail", AssertInputCheckFail(input_bad_kind_config))

		})

		When("Input is valid", func() {
			When("Make directory for kubeconfig_dir with 0700 permission fail", func() {
				p_mkdirall := patchService.AddPatchRequest("os_mkdirall_fail")
				defer patchService.DeletePatchRequest(p_mkdirall)
				It("Should fail", AssertDeployFail(patchService.getContext(), genGoodInput(), output))
			})
			When("Make directory for kubeconfig_dir with 0700 permission ok", func() {
				p_mkdirall := patchService.AddPatchRequest("os_mkdirall_ok")
				defer patchService.DeletePatchRequest(p_mkdirall)
				When("Pull input files from mirrorurl to kindbin ok", func() {

					When("Change kindbin mode to 700 fail", func() {
						srl := patchService.AddPatchRequest("os_chmod_fail")
						defer patchService.DeletePatchRequest(srl)

						input_good := make(map[string][]byte)
						for key, value := range input_template_good {
							input_good[key] = value
						}
						It("Should fail", AssertDeployFail(patchService.getContext(), input_good, output))
					})

					When("Change kindbin mode to 700 ok", func() {
						srl := patchService.AddPatchRequest("os_chmod_ok")
						defer patchService.DeletePatchRequest(srl)
						When("Runtimedata does not exist", func() {
							p1 := patchService.AddPatchRequest("os.Stat_fail")
							defer patchService.DeletePatchRequest(p1)
							p2 := patchService.AddPatchRequest("os.IsNotExist_y")
							defer patchService.DeletePatchRequest(p2)
							When("Make directory for input_ep_params.Runtimedata with permission os.ModePerm fail", func() {
								patchService.DeletePatchRequest(p_mkdirall)
								defer patchService.AddPatchRequest(p_mkdirall)
								srl := patchService.AddPatchRequest("os_mkdirall_1")
								defer patchService.DeletePatchRequest(srl)
								It("Should fail", AssertDeployFail(patchService.getContext(), genGoodInput(), output))
							})
						})
						When("Runtimedata exist", func() {
							p1 := patchService.AddPatchRequest("os.Stat_ok")
							p2 := patchService.AddPatchRequest("os.IsNotExist_n")
							defer patchService.DeletePatchRequest(p1)
							defer patchService.DeletePatchRequest(p2)
							When("Make directory for input_ep_params.Runtimedata with permission os.ModePerm ok", func() {
								patchService.DeletePatchRequest(p_mkdirall)
								defer patchService.AddPatchRequest(p_mkdirall)
								srl := patchService.AddPatchRequest("os_mkdirall_ok")
								defer patchService.DeletePatchRequest(srl)
								When("Write kindcluster config target with input kindconfig with permission 600 fail", func() {
									p1 := patchService.AddPatchRequest("ioutil.WriteFile_fail")
									defer patchService.DeletePatchRequest(p1)
									It("Should fail", AssertDeployFail(patchService.getContext(), genGoodInput(), output))
								})
								When("Write kindcluster config target with input kindconfig with permission 600 ok", func() {
									p1 := patchService.AddPatchRequest("ioutil.WriteFile_ok")
									defer patchService.DeletePatchRequest(p1)
									When("Deploying kind fail", func() {
										p1 := patchService.AddPatchRequest("eputil_runcmdex_fail")
										defer patchService.DeletePatchRequest(p1)
										It("Should fail", AssertDeployFail(patchService.getContext(), genGoodInput(), output))
									})
									When("Deploying kind success", func() {
										srl := patchService.AddPatchRequest("eputil_runcmdex_ok")
										defer patchService.DeletePatchRequest(srl)
										When("Remove kind cluster instance config tgt fail", func() {
											p1 := patchService.AddPatchRequest("eputils.RemoveFile_fail")
											defer patchService.DeletePatchRequest(p1)
											It("Should fail", AssertDeployFail(patchService.getContext(), genGoodInput(), output))
										})
										When("Remove kind cluster instance config tgt ok", func() {
											p1 := patchService.AddPatchRequest("eputils.RemoveFile_ok")
											defer patchService.DeletePatchRequest(p1)
											When("Read kubeconfig by ioutil.readfile fail", func() {
												srl := patchService.AddPatchRequest("ioutil_readfile_fail")
												defer patchService.DeletePatchRequest(srl)
												It("Should fail", AssertDeployFail(patchService.getContext(), genGoodInput(), output))
											})
											When("Read kubeconfig by ioutil.readfile ok", func() {
												srl := patchService.AddPatchRequest("ioutil_readfile_ok")
												defer patchService.DeletePatchRequest(srl)
												When("RuntimeDir is valid", func() {
													p1 := patchService.AddPatchRequest("os.mkdir_ok")
													defer patchService.DeletePatchRequest(p1)
													It("Should ok", AssertDeploySuccess(patchService.getContext(), genGoodInput(), output))
												})
											})
										})

									})

								})

							})
						})
					})
				})

				When("Pull input files from mirrorurl to kindbin fail", func() {
					srl := patchService.AddPatchRequest("repoutils_pullfilefromrepo_fail")
					defer patchService.DeletePatchRequest(srl)

					It("Should fail", AssertDeployFail(patchService.getContext(), genGoodInput(), output))
				})
			})

		})

	})
	//endregion

})
