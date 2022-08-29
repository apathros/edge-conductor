/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package workflow

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	epapiplugins "ep/pkg/api/plugins"
	pluginapi "ep/pkg/api/plugins"
	wfapi "ep/pkg/api/workflow"
	epplugins "ep/pkg/epplugins"
	_ "ep/pkg/epplugins/file-exporter"
	eputils "ep/pkg/eputils"
	docker "ep/pkg/eputils/docker"
	plugin "ep/pkg/plugin"

	mpatch "github.com/undefinedlabs/go-mpatch"
)

var (
	errEmpty = fmt.Errorf("")
	errTest  = fmt.Errorf("test error")
)

var funcs = template.FuncMap{
	"readfile": func(f string) (string, error) {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			return "", nil
		}
		return string(b), nil
	},
}

var server_init = &server{
	workflow: &wfapi.Workflow{
		Spec: &wfapi.WorkflowSpec{
			Data: []*wfapi.WorkflowSpecDataItems0{
				{
					Name:  "test",
					Value: "[]",
				},
			},
			Containers: wfapi.Containers{
				&wfapi.ContainersItems0{
					Name: "test",
				},
			},
		},
	},
	name: "test",
	steps: []step{
		{
			plugin:    "test",
			container: "test",
		},
	},
	plugin_data:      eputils.SchemaMapData{},
	plugin_dataattrs: map[string]dataAttr{},
	finished:         make(chan bool),
	data:             &wfapi.WorkflowData{},
	errch:            make(chan error),
}

func Test_getPendingStep(t *testing.T) {
	s := server_init
	s.getPendingStep("test")
}

func Test_IsConfidentialData(t *testing.T) {
	s := server_init
	s.IsConfidentialData("test")
}

func Test_loadPluginData(t *testing.T) {
	s := server_init
	err := s.loadPluginData()
	require.NoError(t, err, "Load Plugin Data Error:")
}

func Test_loadContainers(t *testing.T) {
	s := server_init
	s.loadContainers()
}

func Test_loadPluginConfig(t *testing.T) {
	s := server_init
	s.loadPluginConfig()
}

type PluginPutLogServer struct {
	wfapi.Workflow_PluginPutLogServer
}

func (*PluginPutLogServer) Recv() (*wfapi.Log, error) {
	return nil, errEmpty
}

func Test_PluginPutLog(t *testing.T) {
	s := server_init
	err := s.PluginPutLog(&PluginPutLogServer{})
	require.NoError(t, err, "Plugin Put Log Error:")
}

var (
	Name   = "test"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
	count  = 0
)

func __name(n string) string {
	return Name + "." + n
}

func test_plugin_main(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	count++
	if count > 3 {
		return errEmpty
	} else {
		return nil
	}
}

func test_plugin_init() {
	eputils.AddSchemaStruct(__name("test-input"), func() eputils.SchemaStruct { return &epapiplugins.Filecontent{} })
	eputils.AddSchemaStruct(__name("test-output"), func() eputils.SchemaStruct { return &epapiplugins.Filecontent{} })

	Input[__name("test-input")] = &epapiplugins.Filecontent{}
	Output[__name("test-output")] = &epapiplugins.Filecontent{}

	plugin.RegisterPlugin(Name, &Input, &Output, test_plugin_main)
}

func Test_Start(t *testing.T) {
	test_plugin_init()

	plugin.Address = "localhost:50089"

	epplugins.PluginList = append(epplugins.PluginList, "test")

	err := Start("testErr", "localhost:100000", "./workflow/workflow.yml")
	t.Log("workflow ret:", err)

	err = Start("test", "localhost:100000", "./workflow/workflow.yml")
	t.Log("workflow ret:", err)

	err = Start("test", "localhost:50089", "./workflow/workflow.yml")
	t.Log("workflow ret:", err)
	t.Log("Done")
}

func unpatch(t *testing.T, m *mpatch.Patch) {
	err := m.Unpatch()
	if err != nil {
		t.Fatal(err)
	}
}

func handle_ctn_ok(*pluginapi.ContainersItems0) error {
	return nil
}

func handle_ctn_err(*pluginapi.ContainersItems0) error {
	return errTest
}

func convert_schema_err(interface{}, interface{}) error {
	return errTest
}

func Test_run_container(t *testing.T) {
	c := &wfapi.ContainersItems0{
		Name:  "test",
		Image: "busybox",
	}

	cases := []struct {
		name               string
		pfunc_dockerrun    func(*pluginapi.ContainersItems0) error
		pfunc_dockerremove func(*pluginapi.ContainersItems0) error
		pfunc_convert      func(interface{}, interface{}) error
		expectError        error
	}{
		{
			name:               "docker_run fail",
			pfunc_dockerremove: handle_ctn_ok,
			pfunc_dockerrun:    handle_ctn_err,
			expectError:        errTest,
		},
		{
			name:               "docker_remove fail",
			pfunc_dockerremove: handle_ctn_err,
			pfunc_dockerrun:    handle_ctn_ok,
			expectError:        nil,
		},
		{
			name:               "convert fail",
			pfunc_dockerremove: handle_ctn_ok,
			pfunc_dockerrun:    handle_ctn_ok,
			pfunc_convert:      convert_schema_err,
			expectError:        eputils.GetError("errConvContainers"),
		},
		{
			name:               "docker ok",
			pfunc_dockerremove: handle_ctn_ok,
			pfunc_dockerrun:    handle_ctn_ok,
			expectError:        nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.pfunc_dockerremove != nil {
				p1, err := mpatch.PatchMethod(docker.DockerRemove, tc.pfunc_dockerremove)
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, p1)
			}
			if tc.pfunc_dockerrun != nil {
				p2, err := mpatch.PatchMethod(docker.DockerRun, tc.pfunc_dockerrun)
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, p2)
			}
			if tc.pfunc_convert != nil {
				p3, err := mpatch.PatchMethod(eputils.ConvertSchemaStruct, tc.pfunc_convert)
				if err != nil {
					t.Fatal(err)
				}
				defer unpatch(t, p3)
			}

			err := run_container(c)

			if tc.expectError == nil && err != nil {
				t.Errorf("unexpected error happened.")
			}
			if tc.expectError != nil &&
				(err == nil || !strings.Contains(err.Error(), tc.expectError.Error())) {
				t.Errorf("Expect error %s but got %s.", tc.expectError, err)
			}
		})
	}
}

func TestMain(m *testing.M) {
	_, pwdpath, _, _ := runtime.Caller(0)
	testdatapath := filepath.Join(filepath.Dir(pwdpath), "testdata")
	workspaceDir, err := os.MkdirTemp("", "workspace")
	if err != nil {
		log.Error(err)
		return
	}
	defer os.RemoveAll(workspaceDir)
	if err := os.Chdir(workspaceDir); err != nil {
		log.Errorf("Failed to change to directory %s", workspaceDir)
		return
	}

	log.Infof("Test workspace: %s\n", workspaceDir)
	cmd := exec.Command("cp", "-a", fmt.Sprintf("%s/workspace/workflow/", testdatapath), fmt.Sprintf("%s/", workspaceDir))
	if _, err := eputils.RunCMD(cmd); err != nil {
		log.Errorf("Failed to run command")
		return
	}
	cmd = exec.Command("cp", "-a", fmt.Sprintf("%s/workspace/runtime/", testdatapath), fmt.Sprintf("%s/", workspaceDir))
	if _, err := eputils.RunCMD(cmd); err != nil {
		log.Errorf("Failed to run command")
		return
	}
	cmd = exec.Command("cp", "-a", fmt.Sprintf("%s/workspace/cert/", testdatapath), fmt.Sprintf("%s/", workspaceDir))
	if _, err := eputils.RunCMD(cmd); err != nil {
		log.Errorf("Failed to run command")
		return
	}

	epparams := &epapiplugins.EpParams{
		Workspace:   workspaceDir,
		Runtimedir:  workspaceDir + "/runtime",
		Runtimebin:  workspaceDir + "/runtime/bin",
		Runtimedata: workspaceDir + "/runtime/data",
	}
	eputils.SetTemplateParams(epparams)
	eputils.SetTemplateFuncs(funcs)

	os.Exit(m.Run())
}
