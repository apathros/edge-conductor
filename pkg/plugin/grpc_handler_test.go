/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package plugin_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	epapiplugins "github.com/intel/edge-conductor/pkg/api/plugins"
	epplugins "github.com/intel/edge-conductor/pkg/epplugins"
	_ "github.com/intel/edge-conductor/pkg/epplugins/file-exporter"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	plugin "github.com/intel/edge-conductor/pkg/plugin"
	wf "github.com/intel/edge-conductor/pkg/workflow"
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
		return eputils.GetError("errEmpty")
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

	plugin.Address = "localhost:50090"

	errtest := plugin.EnablePluginRemoteLog("test")
	require.NoError(t, errtest, "Enable Plugin Remote Log Error:")

	errtest001 := plugin.EnablePluginRemoteLog("test001")
	if errtest001 != nil {
		t.Log("Expected error :", errtest001)
	} else {
		t.Fatal("Unexpected test result")
	}

	epplugins.PluginList = append(epplugins.PluginList, "test")

	err := wf.Start("test", "localhost:100000", "./workflow/workflow.yml")
	t.Log("workflow ret:", err)

	err = wf.Start("test", "localhost:50090", "./workflow/workflow.yml")
	t.Log("workflow ret:", err)

	errWaitPluginFinished := plugin.WaitPluginFinished("file-exporter")
	if errWaitPluginFinished != nil {
		t.Fatal("Expected error :", errWaitPluginFinished)
	}
	errWaitPluginFinished01 := plugin.WaitPluginFinished("test001")
	if errWaitPluginFinished01 != nil {
		t.Log("Expected error :", errWaitPluginFinished01)
	} else {
		t.Fatal("Unexpected test result")
	}
	t.Log("Done")
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
