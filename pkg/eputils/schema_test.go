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

	"os"
	"testing"
	"text/template"
)

func TestSchema(t *testing.T) {
	var stryml string
	var err error
	saveschemafile := "testdata/schema_test_save.yml"
	epparams := &pluginapi.EpParams{}
	epparams2 := &pluginapi.EpParams{}

	t.Log("SaveSchemaStructToYamlFile")
	if err = SaveSchemaStructToYamlFile(epparams, saveschemafile); err != nil {
		t.Error(err)
	}
	t.Log("LoadSchemaStructFromYamlFile ok")
	if err = LoadSchemaStructFromYamlFile(epparams, saveschemafile); err != nil {
		t.Error(err)
	}
	t.Log("LoadSchemaStructFromYamlFile err")
	if err = LoadSchemaStructFromYamlFile(epparams, "testdata/schema_test_invalid_yaml.yml"); err == nil {
		t.Error("Expect error but not found")
	}
	t.Log("SchemaStructToYaml")
	if stryml, err = SchemaStructToYaml(epparams); err != nil {
		t.Error(err)
	}
	t.Log("LoadSchemaStructFromYaml")
	if err = LoadSchemaStructFromYaml(epparams, stryml); err != nil {
		t.Error(err)
	}
	t.Log("ConvertSchemaStruct")
	if err = ConvertSchemaStruct(epparams, epparams2); err != nil {
		t.Error(err)
	}
	t.Log("LoadMapFromYamlFile ok")
	ymlmap, err := LoadMapFromYamlFile(saveschemafile)
	if err != nil {
		t.Error(err)
	}
	t.Log("ConvertStructToMap ok")
	ymlmap2, err := ConvertStructToMap(epparams2)
	if err != nil {
		t.Error(err)
	}
	t.Log("MergeMaps ok")
	ymlmap3 := MergeMaps(ymlmap, ymlmap2)
	t.Log(ymlmap3)

	t.Log("FileTemplateConvert")
	if err = FileTemplateConvert(saveschemafile, saveschemafile); err != nil {
		t.Error(err)
	}

	t.Log("StringTemplateConvert")
	newstr, err := StringTemplateConvert("test: template")
	if err != nil {
		t.Error(err)
	}
	t.Log(newstr)

	t.Log("StringTemplateConvertWithParams")
	newstr2, err := StringTemplateConvertWithParams("test: template", epparams)
	if err != nil {
		t.Error(err)
	}
	t.Log(newstr2)

	t.Log("LoadJsonFile ok")
	jsondata, err := LoadJsonFile(saveschemafile)
	if err != nil {
		t.Error(err)
	}
	t.Log(jsondata)

	emptybin := []byte(`{}`)

	t.Log("SchemaMapData")
	newmapdata := NewSchemaMapData()
	newmapdata2 := NewSchemaMapData()
	if err = newmapdata.UnmarshalBinary(emptybin); err != nil {
		t.Error(err)
	}
	if err = newmapdata2.UnmarshalBinary(emptybin); err != nil {
		t.Error(err)
	}
	if !newmapdata.EqualWith(newmapdata2) {
		t.Error("Data should be equal.")
	}
	_, err = newmapdata.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if err = newmapdata.Validate(nil); err != nil {
		t.Error(err)
	}

	t.Log("SchemaData")
	newdata := &SchemaData{}
	newdata2 := &SchemaData{}
	if err = newdata.UnmarshalBinary(emptybin); err != nil {
		t.Error(err)
	}
	if err = newdata2.UnmarshalBinary(emptybin); err != nil {
		t.Error(err)
	}
	_, err = newdata.MarshalBinary()
	if err != nil {
		t.Error(err)
	}
	if err = newdata.Validate(nil); err != nil {
		t.Error(err)
	}

	t.Log("SchemaStruct")
	schemastr := SchemaStructNew("test")
	if schemastr == nil {
		t.Error("schemastr is nil")
	}
	SetTemplateParams(schemastr)
	SetTemplateFuncs(template.FuncMap{})

	// Cleanup
	os.RemoveAll(saveschemafile)
}
