/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//nolint: dupl
package restfulcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/intel/edge-conductor/pkg/eputils"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errCliTest    = errors.New("test_error")
	errClientPost = errors.New("client post error")
)

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func TestTlsBasicAuth(t *testing.T) {
	cases := []struct {
		name           string
		user           string
		password       string
		expectedOutput string
	}{
		{
			name:           "Test Tls Basic Authorization",
			user:           "test",
			password:       "test123",
			expectedOutput: "basic dGVzdDp0ZXN0MTIz",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := TlsBasicAuth(tc.user, tc.password)
			if output != tc.expectedOutput {
				t.Errorf("Unexpected return value: %v, expected: %v", output, tc.expectedOutput)
			}
		})
	}
}

/*
 Responsed body of "GET" https://harbor/api/v2.0/projects/library

 success:
  {"chart_count":0,"creation_time":"2022-01-07T09:24:04.787Z","current_user_role_id":1,"current_user_role_ids":[1],"cve_allowlist":{"creation_time":"0001-01-01T00:00:00.000Z","id":1,"items":[],"project_id":1,"update_time":"0001-01-01T00:00:00.000Z"},"metadata":{"public":"true"},"name":"library","owner_id":1,"owner_name":"admin","project_id":1,"repo_count":4,"update_time":"2022-01-07T09:24:04.787Z"}


 failed:
  {"errors":[{"code":"FORBIDDEN","message":"forbidden"}]}

 Use go-mpatch to mask resty function, and assign value to the result after resty function.
*/

type testJsonStructInfo struct {
	KeyInfo string `json:"info"`
}
type testJsonStructFail struct {
	KeyFail       bool
	KeyStatusCode string `json:"StatusCode"`
}
type testJsonStruct struct {
	testJsonStructInfo
	testJsonStructFail
}

func (t1 testJsonStructInfo) printJson1() testJsonStructInfo {
	return t1
}

func (t2 testJsonStructFail) printJson2() testJsonStructFail {
	return t2
}

func TestRegistryProjectExists(t *testing.T) {
	cases := []struct {
		name           string
		registry       string
		project        string
		authString     string
		certFilePath   string
		ResponseBody   string
		expectedOutput testJsonStruct
	}{
		{
			name:           "Test getting project successfully ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "name:test",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Project exist"}, testJsonStructFail{KeyFail: false, KeyStatusCode: "200"}},
		},
		{
			name:           "Test getting project failed ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Project not found"}, testJsonStructFail{KeyFail: false, KeyStatusCode: "404"}},
		},
		{
			name:           "Test getting project failed ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Harbor response is abnormal"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test getting project failed ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "client get error"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without registry ",
			registry:       "",
			project:        "testProject",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Harbor URL string is null"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without project ",
			registry:       "192.168.250.100",
			project:        "",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Project string is null"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without authString ",
			registry:       "192.168.250.100",
			project:        "testProject",
			authString:     "",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "auth string is null"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without File Path ",
			registry:       "192.168.250.100",
			project:        "testProject",
			authString:     "testAuth",
			certFilePath:   "",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Cert file is null"}, testJsonStructFail{KeyFail: true}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r *resty.Request
			var respFail *resty.Response
			respSuccess := new(resty.Response)

			patch1, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(r), "Get", func(r *resty.Request, a string) (*resty.Response, error) {
				if !tc.expectedOutput.printJson2().KeyFail {
					return respSuccess, nil
				} else {
					if tc.expectedOutput.printJson1().KeyInfo == "client get error" {
						return respFail, eputils.GetError("errClientGet")
					} else {
						return respFail, nil
					}
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch1)

			patch2, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(respSuccess), "StatusCode", func(respSuccess *resty.Response) int {
				if !tc.expectedOutput.printJson2().KeyFail {
					if tc.expectedOutput.printJson2().KeyStatusCode != "" {
						statusCode, err := strconv.Atoi(tc.expectedOutput.printJson2().KeyStatusCode)
						if err != nil {
							t.Fatal("can't convert to int")
						}
						return statusCode
					} else {
						return 500
					}
				}
				return 500
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch2)

			output, _ := RegistryProjectExists(tc.registry, tc.project, tc.authString, tc.certFilePath)
			if tc.name == "Test getting project successfully " {
				if !output {
					t.Errorf("failed")
				}
			} else {
				if output {
					t.Errorf("failed")
				}
			}
		})
	}
}

func TestRegistryCreateProject(t *testing.T) {
	cases := []struct {
		name           string
		registry       string
		project        string
		authString     string
		certFilePath   string
		ResponseBody   string
		expectedOutput testJsonStruct
	}{
		{
			name:           "Test create project successfully ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "name:test",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Create project successfully"}, testJsonStructFail{KeyFail: false, KeyStatusCode: "201"}},
		},
		{
			name:           "Test create project exist ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "name:test",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Project already exist"}, testJsonStructFail{KeyFail: false, KeyStatusCode: "409"}},
		},
		{
			name:           "Test create project failed ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Harbor response error"}, testJsonStructFail{KeyFail: false, KeyStatusCode: "500"}},
		},
		{
			name:           "Test create project failed ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Harbor response is abnormal"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test create project failed ",
			registry:       "192.168.250.100",
			project:        "test",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "client post error"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without File Path ",
			registry:       "192.168.250.100",
			project:        "testProject",
			authString:     "testAuth",
			certFilePath:   "",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "cert file is null"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without authString ",
			registry:       "192.168.250.100",
			project:        "testProject",
			authString:     "",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "auth string is null"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without registry ",
			registry:       "",
			project:        "testProject",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Harbor URL string is null"}, testJsonStructFail{KeyFail: true}},
		},
		{
			name:           "Test without project ",
			registry:       "192.168.250.100",
			project:        "",
			authString:     "testAuth",
			certFilePath:   "./testdata/testca.crt",
			ResponseBody:   "{error:errortest}",
			expectedOutput: testJsonStruct{testJsonStructInfo{KeyInfo: "Project string is null"}, testJsonStructFail{KeyFail: true}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var r *resty.Request
			var respFail *resty.Response
			respSuccess := new(resty.Response)

			patch1, err := mpatch.PatchInstanceMethodByName(reflect.TypeOf(r), "Post", func(r *resty.Request, a string) (*resty.Response, error) {
				if !tc.expectedOutput.printJson2().KeyFail {
					return respSuccess, nil
				} else {
					if tc.expectedOutput.printJson1().KeyInfo == "client post error" {
						return respFail, errClientPost
					} else {
						return respFail, nil
					}
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch1)
			patch2, err := mpatch.PatchMethod(fmt.Sprintf, func(string, ...interface{}) string {
				if !tc.expectedOutput.printJson2().KeyFail {
					if tc.expectedOutput.printJson2().KeyStatusCode != "" {
						return tc.expectedOutput.printJson2().KeyStatusCode
					} else {
						encodedBody, _ := json.Marshal(tc.expectedOutput.printJson2())
						encodedBodyStringType := string(encodedBody)
						return encodedBodyStringType
					}
				} else {
					encodedBody, _ := json.Marshal(tc.expectedOutput.printJson1())
					encodedBodyStringType := string(encodedBody)
					return encodedBodyStringType
				}
			})
			if err != nil {
				t.Fatal(err)
			}
			defer unpatch(t, patch2)
			output := RegistryCreateProject(tc.registry, tc.project, tc.authString, tc.certFilePath)
			if tc.name == "Test create project successfully " || tc.name == "Test create project exist " {
				if output != nil {
					t.Errorf("failed")
				}
			} else {
				if output == nil {
					t.Errorf("failed")
				}
			}
		})
	}
}

func TestMapImageURLOnHarbor(t *testing.T) {
	cases := []struct {
		name             string
		in_image         string
		expect_out_image string
		expect_err       bool
	}{
		{
			name:             "docker.io_layer1",
			in_image:         "nginx:latest",
			expect_out_image: "docker.io/library/nginx:latest",
			expect_err:       false,
		},
		{
			name:             "docker.io_layer2",
			in_image:         "kindest/node:latest",
			expect_out_image: "docker.io/kindest/node:latest",
			expect_err:       false,
		},
		{
			name:             "non_docker.io",
			in_image:         "k8s.gcr.io/pause:latest",
			expect_out_image: "k8s.gcr.io/pause:latest",
			expect_err:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			images := []string{tc.in_image}
			out_images, err := MapImageURLOnHarbor(images)
			if !tc.expect_err {
				if err != nil {
					t.Errorf("Unexpected error.")
				} else {
					newImage := out_images[0]
					if newImage != tc.expect_out_image {
						t.Errorf("Input %s, expected %s, but got %s.", tc.in_image, tc.expect_out_image, newImage)
					}
				}
			} else {
				if err == nil {
					t.Errorf("Should report error but no error found.")
				}
			}

		})
	}
}

func TestCreateHarborProject(t *testing.T) {
	func_RegistryProjectExists_fail := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRegistryProjectExists, err := mpatch.PatchMethod(RegistryProjectExists, func(harborUrl string, project string, authStr string, certFilePath string) (bool, error) {
			return false, errCliTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchRegistryProjectExists}
	}
	func_RegistryProject_Not_Exists := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRegistryProjectExists, err := mpatch.PatchMethod(RegistryProjectExists, func(harborUrl string, project string, authStr string, certFilePath string) (bool, error) {
			return false, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchRegistryCreateProject, err := mpatch.PatchMethod(RegistryCreateProject, func(harborUrl string, project string, authStr string, certFilePath string) error {
			return errCliTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchRegistryProjectExists, pathchRegistryCreateProject}
	}
	func_runnormal_ok := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRegistryProjectExists, err := mpatch.PatchMethod(RegistryProjectExists, func(harborUrl string, project string, authStr string, certFilePath string) (bool, error) {
			return false, nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}
		pathchRegistryCreateProject, err := mpatch.PatchMethod(RegistryCreateProject, func(harborUrl string, project string, authStr string, certFilePath string) error {
			return nil
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchRegistryProjectExists, pathchRegistryCreateProject}
	}
	type args struct {
		authServerAddress   string
		projectName         string
		authStr             string
		DayZeroCertFilePath string
	}
	cases := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add test cases.
		{
			name: "authServerAddress_nil",
			args: args{
				authServerAddress:   "",
				projectName:         "",
				authStr:             "",
				DayZeroCertFilePath: "",
			},
			expectErrorContent: eputils.GetError("errInputAuthSrv"),
		},
		{
			name: "projectName_nil",
			args: args{
				authServerAddress:   "deafalut_addr",
				projectName:         "",
				authStr:             "",
				DayZeroCertFilePath: "",
			},
			expectErrorContent: eputils.GetError("errProjectName"),
		},
		{
			name: "authString_nil",
			args: args{
				authServerAddress:   "deafalut_addr",
				projectName:         "deafalut_proName",
				authStr:             "",
				DayZeroCertFilePath: "",
			},
			expectErrorContent: eputils.GetError("errAuthEmpty"),
		},
		{
			name: "DayZeroCertFilePath_nil",
			args: args{
				authServerAddress:   "deafalut_addr",
				projectName:         "deafalut_proName",
				authStr:             "defalut_authStrs",
				DayZeroCertFilePath: "",
			},
			expectErrorContent: eputils.GetError("errCertNull"),
		},
		{
			name: "RegistryProjectExists_err",
			args: args{
				authServerAddress:   "deafalut_addr",
				projectName:         "deafalut_proName",
				authStr:             "defalut_authStrs",
				DayZeroCertFilePath: "defalut_dayZeroCertFilePath",
			},
			expectErrorContent: errCliTest,
			funcBeforeTest:     func_RegistryProjectExists_fail,
		},
		{
			name: "RegistryProject_Not_Exists",
			args: args{
				authServerAddress:   "deafalut_addr",
				projectName:         "deafalut_proName",
				authStr:             "defalut_authStrs",
				DayZeroCertFilePath: "defalut_dayZeroCertFilePath",
			},
			expectErrorContent: errCliTest,
			funcBeforeTest:     func_RegistryProject_Not_Exists,
		},
		{
			name: "runnormal_ok",
			args: args{
				authServerAddress:   "deafalut_addr",
				projectName:         "deafalut_proName",
				authStr:             "defalut_authStrs",
				DayZeroCertFilePath: "defalut_dayZeroCertFilePath",
			},
			funcBeforeTest: func_runnormal_ok,
		},
	}
	for _, tc := range cases {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}
		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			if result := CreateHarborProject(tc.args.authServerAddress, tc.args.projectName, tc.args.authStr, tc.args.DayZeroCertFilePath); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}
}

func TestMapImageURLCreateHarborProject(t *testing.T) {
	func_MapImageURLOnHarbor_err := func(ctrl *gomock.Controller) []*mpatch.Patch {
		pathchRegistryProjectExists, err := mpatch.PatchMethod(MapImageURLOnHarbor, func(image []string) ([]string, error) {
			return nil, errCliTest
		})
		if err != nil {
			t.Errorf("patch error: %v", err)
		}

		return []*mpatch.Patch{pathchRegistryProjectExists}
	}
	type args struct {
		harborIP   string
		harborPort string
		harborUser string
		harborPass string
		image      []string
	}
	tests := []struct {
		name               string
		args               args
		expectErrorContent error
		funcBeforeTest     func(*gomock.Controller) []*mpatch.Patch
	}{
		// TODO: Add test cases.
		{
			name: "input_harbor_IP_nil",
			args: args{
				harborIP:   "",
				harborPort: "",
				harborUser: "",
				harborPass: "",
				image:      nil,
			},
			expectErrorContent: eputils.GetError("errHarborIPEmpty"),
		},
		{
			name: "input_harbor_port_nil",
			args: args{
				harborIP:   "10.10.10.10",
				harborPort: "",
				harborUser: "",
				harborPass: "",
				image:      nil,
			},
			expectErrorContent: eputils.GetError("errHarborPort"),
		},
		{
			name: "input_harbor_user_nil",
			args: args{
				harborIP:   "10.10.10.10",
				harborPort: "8080",
				harborUser: "",
				harborPass: "",
				image:      nil,
			},
			expectErrorContent: eputils.GetError("errHarborUser"),
		},
		{
			name: "input_harbor_password_nil",
			args: args{
				harborIP:   "10.10.10.10",
				harborPort: "8080",
				harborUser: "testName",
				harborPass: "",
				image:      nil,
			},
			expectErrorContent: eputils.GetError("errHarborPasswd"),
		},
		{
			name: "MapImageURLOnHarbor_err",
			args: args{
				harborIP:   "10.10.10.10",
				harborPort: "8080",
				harborUser: "testName",
				harborPass: "defalutPwd",
				image:      nil,
			},
			expectErrorContent: errCliTest,
			funcBeforeTest:     func_MapImageURLOnHarbor_err,
		},
	}
	for _, tc := range tests {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		var plist []*mpatch.Patch
		if tc.funcBeforeTest != nil {
			plist = tc.funcBeforeTest(ctrl)
		}
		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			if _, result := MapImageURLCreateHarborProject(tc.args.harborIP, tc.args.harborPort, tc.args.harborUser, tc.args.harborPass, tc.args.image); result != nil {
				if fmt.Sprint(result) == fmt.Sprint(tc.expectErrorContent) {
					t.Log(tc.name, "Done")
				} else {
					t.Errorf("%s error %s", tc.name, result)
				}
			}
		})

		for _, p := range plist {
			unpatch(t, p)
		}
	}
}
