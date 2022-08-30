/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"

	wfapi "github.com/intel/edge-conductor/pkg/api/workflow"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
)

const (
	plugins_folder    = "pkg/epplugins/"
	plugins_yaml_file = "pkg/epplugins/plugins.yml"
)

type schemaStruct struct {
	Name, NameString, SchemaName string
}

type pluginStruct struct {
	PluginName, PluginPkgName string
	InputList, OutputList     []schemaStruct
	IsSchemaAPINeeded         bool
}

/* Define Templates */

// Template for plugin's main.go.
const template_plugin_main_go = `/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package {{.PluginPkgName}}

import (
	// TODO: Add Plugin Imports Here
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	log "github.com/sirupsen/logrus"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
{{- block "ilist1" .}}{{range .InputList}}
	input_{{.Name}} := input_{{.Name}}(in)
{{- end -}}{{- end}}
{{block "olist1" .}}{{range .OutputList}}
	output_{{.Name}} := output_{{.Name}}(outp)
{{- end -}}{{- end}}

	// TODO: Add Plugin Code Here
	log.Infof("Plugin: {{.PluginName}}")
{{- block "ilist2" .}}{{range .InputList}}
	log.Infof("%v", input_{{.Name}})
{{- end -}}{{- end}}
{{block "olist2" .}}{{range .OutputList}}
	log.Infof("%v", output_{{.Name}})
{{- end -}}{{- end}}

	return nil
}
`

// Template for plugin's generated.go.
const template_plugin_generated_go = `/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package {{.PluginPkgName}}

import (
{{ if .IsSchemaAPINeeded }}
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
{{ end -}}
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	epplugin "github.com/intel/edge-conductor/pkg/plugin"
)

var (
	Name = "{{.PluginName}}"
	Input = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}
{{block "ilistfunc" .}}{{range .InputList}}
//nolint:deadcode,unused
func input_{{.Name}}(in eputils.SchemaMapData) *pluginapi.{{.SchemaName}} {
	return in[__name("{{.NameString}}")].(*pluginapi.{{.SchemaName}})
}
{{end}}{{- end}}
{{- block "olistfunc" .}}{{range .OutputList}}
//nolint:deadcode,unused
func output_{{.Name}}(outp *eputils.SchemaMapData) *pluginapi.{{.SchemaName}} {
	return (*outp)[__name("{{.NameString}}")].(*pluginapi.{{.SchemaName}})
}
{{end}}{{- end}}
func init() {
{{- block "ilist1" .}}{{range .InputList}}
	eputils.AddSchemaStruct(__name("{{.NameString}}"), func() eputils.SchemaStruct { return &pluginapi.{{.SchemaName}}{} })
{{- end -}}{{- end -}}
{{- block "olist1" .}}{{range .OutputList}}
	eputils.AddSchemaStruct(__name("{{.NameString}}"), func() eputils.SchemaStruct { return &pluginapi.{{.SchemaName}}{} })
{{- end}}{{- end}}
{{block "ilist2" .}}{{range .InputList}}
	Input[__name("{{.NameString}}")] = &pluginapi.{{.SchemaName}}{}
{{- end -}}{{- end -}}
{{- block "olist2" .}}{{range .OutputList}}
	Output[__name("{{.NameString}}")] = &pluginapi.{{.SchemaName}}{}
{{- end}}{{- end}}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
`

// Template for plugin's main_test.go.
const template_plugin_main_test_go = `/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

//nolint: dupl
package {{.PluginPkgName}}

import (
	"testing"
	// TODO: Add Plugin Unit Test Imports Here
)

func TestPluginMain(t *testing.T) {
	cases := []struct {
		name string
		input, expectedOutput map[string][]byte
		expectError bool
	}{
		// TODO: Add the values to complete your test cases.
		// Add the values for input and expectedoutput with particular struct marshal data in json format.
		// They will be used to generate "SchemaMapData" as inputs and expected outputs of plugins under test.
		// if the inputs in the Plugin Input List is not required in your test case, keep the value as nil.
		{
			name: "CASE/001",
			input: map[string][]byte{
{{- block "ilistdata1" .}}{{range .InputList}}
				"{{.NameString}}": nil,
{{- end -}}{{- end}}
			},
{{block "oexpect1" .}}{{if .OutputList}}
			expectedOutput:  map[string][]byte{
{{block "olistdata1" .}}{{range .OutputList}}
				"{{.NameString}}": nil,
{{- end -}}{{- end}}
			},
{{- end -}}{{- end}}
			expectError: false,
		},

		{
			name: "CASE/002",
			input: map[string][]byte{
{{- block "ilistdata2" .}}{{range .InputList}}
				"{{.NameString}}": nil,
{{- end -}}{{- end}}
			},
{{block "oexpect2" .}}{{if .OutputList}}
			expectedOutput:  map[string][]byte{
{{block "olistdata2" .}}{{range .OutputList}}
				"{{.NameString}}": nil,
{{- end -}}{{- end}}
			},
{{- end -}}{{- end}}
			expectError: true,
		},
	}

	// Optional: add setup for the test series
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Run test cases in parallel if necessary.
			// t.Parallel()

			input := generateInput(tc.input)
			if input == nil {
				t.Fatalf("Failed to generateInput %s", tc.input)
			}
			testOutput := generateOutput(nil)
{{block "conditioncheckprepare" .}}{{if .OutputList}}
			expectedOutput := generateOutput(tc.expectedOutput)
{{- end -}}{{- end}}

			// TODO: Remove the '//' before following check condition to enable plugin test.
			// if result := PluginMain(input, &testOutput); result != nil {
			// 	if tc.expectError {
			// 		t.Log("Error expected.")
			// 		return
			// 	} else {
			// 		t.Logf("Failed to run PluginMain when input is %s.", tc.input)
			// 		t.Error(result)
			// 	}
			// }

{{block "conditioncheck1" .}}{{if .OutputList}}
			if testOutput.EqualWith(expectedOutput) {
				t.Log("Output expected.")
			} else {
				t.Errorf("Failed to get expected output when input is %s.", tc.input)
			}

			// Optional: Add additional check conditions here

{{else}}
			_ = testOutput
			// TODO: Add check conditions
{{- end -}}{{- end}}
		})
	}

	// Optional: add teardown for the test series
}
`

// Template for plugin's generated_testutil.go.
const template_plugin_generated_testutil_go = `/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package {{.PluginPkgName}}

import (
{{ if .IsSchemaAPINeeded }}
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
{{ end -}}
	eputils "github.com/intel/edge-conductor/pkg/eputils"
)
{{block "ilistfunc" .}}{{range .InputList}}
//nolint:deadcode,unused
func generate_input_{{.Name}}(data []byte, in eputils.SchemaMapData) bool {
	inputStruct := &pluginapi.{{.SchemaName}}{}
	if data != nil {
		if err := inputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	in[__name("{{.NameString}}")] = inputStruct
	return true
}
{{end}}{{- end -}}

//nolint:deadcode,unused,unparam
func generateInput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()

{{- block "ilistdata" .}}{{range .InputList}}
	if result := generate_input_{{.Name}}(data["{{.NameString}}"], n); !result {
		return nil
	}
{{- end}}{{- end}}
	return n
}

{{block "olistfunc" .}}{{range .OutputList}}
//nolint:deadcode,unused
func generate_output_{{.Name}}(data []byte, out eputils.SchemaMapData) bool {
	outputStruct := &pluginapi.{{.SchemaName}}{}
	if data != nil {
		if err := outputStruct.UnmarshalBinary(data); err != nil {
			return false
		}
	}

	out[__name("{{.NameString}}")] = outputStruct
	return true
}
{{end}}{{- end -}}

//nolint:unparam,deadcode,unused
func generateOutput(data map[string][]byte) eputils.SchemaMapData {
	n := eputils.NewSchemaMapData()

{{- block "olistdata" .}}{{range .OutputList}}
	if result := generate_output_{{.Name}}(data["{{.NameString}}"], n); !result {
		return nil
	}
{{- end}}{{- end}}
	return n
}
`

// Template for generated.go as a Plugin Index
const template_generated_go = `/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package epplugins

import (
{{- block "list" .}}{{range .}}
	_ "github.com/intel/edge-conductor/pkg/epplugins/{{.PluginName}}"
{{- end -}}{{- end}}
)
var PluginList[]string = []string {
{{- block "list2" .}}{{range .}}
	"{{.PluginName}}",
{{- end -}}{{- end}}
}
`

func MakeDir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func RemoveFile(name string) error {
	err := os.Remove(name)
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
		return err
	}
	return nil
}

func GenPluginGoFile(filename string, plugin pluginStruct, t *template.Template) error {
	if err := RemoveFile(filename); err != nil {
		return err
	}
	f, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
		return err
	}
	if err := t.Execute(f, plugin); err != nil {
		log.Fatal(err)
		return err
	}
	log.Infof("  %s generated.", filename)
	return nil
}

func checkSchema(schemafile string) {
	if strings.Contains(schemafile, "/schemas/") {
		schemafile = strings.Replace(schemafile, "/schemas/", "/.schemas/", 1)
	}
	if !eputils.FileExists(schemafile) {
		log.Errorln("Schema not found:", schemafile)
		log.Fatal("Schema not found!")
	}
	valuemap := map[string]map[string]interface{}{}
	valuebytes, err := ioutil.ReadFile(schemafile)
	if err != nil {
		log.Errorln("Failed to read", schemafile)
		log.Fatal("Schema not available!")
	}
	if err = yaml.Unmarshal(valuebytes, &valuemap); err != nil {
		log.Errorln("Failed to unmarshal", schemafile)
		log.Fatal("Schema not available!")
	}

	schemaname := strings.TrimSuffix(filepath.Base(schemafile), filepath.Ext(schemafile))
	for k, _ := range valuemap["definitions"] {
		if k != schemaname {
			log.Errorln("Schema name", k, "do not match file name", schemafile)
			log.Fatal("Schema name mis-match.")
		}
	}
}

func main() {
	tpmain := template.Must(template.New("pmaingo").Parse(template_plugin_main_go))
	tpgenerated := template.Must(template.New("pgeneratedgo").Parse(template_plugin_generated_go))
	tpmaintest := template.Must(template.New("pmaintestgo").Parse(template_plugin_main_test_go))
	tptestutil := template.Must(template.New("pgeneratedtestgo").Parse(template_plugin_generated_testutil_go))

	tgenerated := template.Must(template.New("generatedgo").Parse(template_generated_go))

	var plugins []pluginStruct

	pluginsobj := wfapi.Plugins{}
	err := eputils.LoadSchemaStructFromYamlFile(&pluginsobj, plugins_yaml_file)
	if err != nil {
		log.Fatal(err)
	}

	for _, plugin := range pluginsobj.Plugins {
		pluginfolder := filepath.Join(plugins_folder, plugin.Name)
		pluginmaingo := filepath.Join(pluginfolder, "main.go")
		plugingeneratedgo := filepath.Join(pluginfolder, "generated.go")
		pluginmaintestgo := filepath.Join(pluginfolder, "main_test.go")
		plugintestutilgo := filepath.Join(pluginfolder, "generated_test.go")

		log.Infof("Plugin Name: %s", plugin.Name)

		var thisplugin pluginStruct
		thisplugin.PluginName = plugin.Name
		thisplugin.PluginPkgName = strings.ReplaceAll(plugin.Name, "-", "")
		for _, in := range plugin.Input {
			checkSchema(in.Schema)
			schemaFileName := filepath.Base(in.Schema)
			thisplugin.InputList = append(thisplugin.InputList,
				schemaStruct{
					strings.ReplaceAll(in.Name, "-", "_"),
					in.Name,
					strings.ReplaceAll(strings.Title(strings.TrimSuffix(schemaFileName, filepath.Ext(schemaFileName))), "-", "")})
		}
		for _, out := range plugin.Output {
			checkSchema(out.Schema)
			schemaFileName := filepath.Base(out.Schema)
			thisplugin.OutputList = append(thisplugin.OutputList,
				schemaStruct{
					strings.ReplaceAll(out.Name, "-", "_"),
					out.Name,
					strings.ReplaceAll(strings.Title(strings.TrimSuffix(schemaFileName, filepath.Ext(schemaFileName))), "-", "")})
		}
		if len(thisplugin.InputList)+len(thisplugin.OutputList) == 0 {
			thisplugin.IsSchemaAPINeeded = false
		} else {
			thisplugin.IsSchemaAPINeeded = true
		}

		plugins = append(plugins, thisplugin)

		if !FileExists(pluginfolder) {
			if err := MakeDir(pluginfolder); err != nil {
				log.Fatal(err)
				return
			}
			log.Infof("  %s created.", pluginfolder)
		}

		// generated*.go will always be re-generated.
		if err := GenPluginGoFile(plugingeneratedgo, thisplugin, tpgenerated); err != nil {
			log.Fatal(err)
			return
		}
		if err := GenPluginGoFile(plugintestutilgo, thisplugin, tptestutil); err != nil {
			log.Fatal(err)
			return
		}
		// main*.go will only be generated once, as these files are maintained by plugin owners.
		if !FileExists(pluginmaingo) {
			if err := GenPluginGoFile(pluginmaingo, thisplugin, tpmain); err != nil {
				log.Fatal(err)
				return
			}
		}
		if !FileExists(pluginmaintestgo) {
			if err := GenPluginGoFile(pluginmaintestgo, thisplugin, tpmaintest); err != nil {
				log.Fatal(err)
				return
			}
		}

	}

	generatedgo := filepath.Join(plugins_folder, "generated.go")
	if err := RemoveFile(generatedgo); err != nil {
		log.Fatal(err)
		return
	}
	f, err := os.Create(generatedgo)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = tgenerated.Execute(f, plugins)
	if err != nil {
		log.Fatal(err)
	}

}
