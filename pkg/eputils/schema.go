/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

//go:generate mockgen -destination=./mock/schemautil_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt github.com/intel/edge-conductor/pkg/eputils SchemaWrapper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

type SchemaWrapper interface {
	LoadJsonFile(file string) ([]byte, error)
}

type SchemaStruct interface {
	Validate(formats strfmt.Registry) error
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(b []byte) error
}

var (
	schemaStructNew map[string]func() SchemaStruct = map[string]func() SchemaStruct{}
	templateParams  interface{}                    = map[string]string{}
	templateFuncs                                  = sprig.TxtFuncMap()
)

func AddSchemaStruct(name string, newFunc func() SchemaStruct) {
	schemaStructNew[name] = newFunc
}

func SchemaStructNew(name string) SchemaStruct {
	if _, has := schemaStructNew[name]; !has {
		log.Infof("Cannot find schema name: %v, set it to interface{}", name)
		AddSchemaStruct(name, func() SchemaStruct { return &SchemaData{} })
	}
	return schemaStructNew[name]()
}

/*
 var SchemaStructNew map[string] func () SchemaStruct = map[string] func () SchemaStruct {
	 "images": func () SchemaStruct { return &pluginapi.Images{} },
	 "nodes": func () SchemaStruct { return &pluginapi.Nodes{} },
	 "sample": func () SchemaStruct { return &pluginapi.Pluginsample{} },
 }
*/

type SchemaMapData map[string]SchemaStruct

func SetTemplateParams(p interface{}) {
	templateParams = p
}

func SetTemplateFuncs(funcs template.FuncMap) {
	for k, v := range funcs {
		templateFuncs[k] = v
	}
}

func NewSchemaMapData() SchemaMapData {
	return make(map[string]SchemaStruct)
}

func (m *SchemaMapData) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

func (m *SchemaMapData) UnmarshalBinary(b []byte) error {
	d := &map[string]interface{}{}

	if err := json.Unmarshal(b, d); err != nil {
		return err
	}
	n := NewSchemaMapData()
	for k, v := range *d {
		vjson, err := json.Marshal(v)
		if err != nil {
			return err
		}
		n[k] = SchemaStructNew(k)
		if err := n[k].UnmarshalBinary(vjson); err != nil {
			return err
		}
	}
	*m = n

	return nil
}

func (m *SchemaMapData) Validate(formats strfmt.Registry) error {
	var res []error

	for _, v := range *m {
		if err := v.Validate(formats); err != nil {
			res = append(res, err)
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

type SchemaData map[string]interface{}

func (m *SchemaData) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

func (m *SchemaData) UnmarshalBinary(b []byte) error {
	if err := json.Unmarshal(b, m); err != nil {
		return err
	}
	return nil
}

func (m *SchemaData) Validate(formats strfmt.Registry) error {
	return nil
}

/*
 func (m *SchemaMapData) Merge(a *SchemaMapData) error {
 }
*/

func (m *SchemaMapData) EqualWith(compare SchemaMapData) bool {
	binaryOri, errOri := (*m).MarshalBinary()
	binaryCom, errCom := compare.MarshalBinary()

	if errOri != nil || errCom != nil {
		log.Errorf("MarshalBinary error: errOri: %v, errCom: %v", errOri, errCom)
		return false
	}

	if res := bytes.Compare(binaryOri, binaryCom); res != 0 {
		log.Debugf("binary is: %s and the binary of compare is: %s", binaryOri, binaryCom)
		return false
	}
	return true
}

func LoadJsonFile(file string) ([]byte, error) {
	src, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = CheckHashForContent(src, file, "")
	if err != nil {
		if err.Error() == ERRORCODE_CHECK_HASH_FAIL {
			log.Warnln("Failed to check file hash:", err)
		} else {
			return nil, err
		}

	}

	tpl := template.Must(template.New(file).Funcs(templateFuncs).Parse(string(src)))
	var b bytes.Buffer
	if err := tpl.Execute(&b, templateParams); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func StringTemplateConvert(str string) (string, error) {
	tpl := template.Must(template.New("StringTemplateConvert").Funcs(templateFuncs).Parse(str))
	var b bytes.Buffer
	if err := tpl.Execute(&b, templateParams); err != nil {
		return "", err
	}
	return b.String(), nil
}

func StringTemplateConvertWithParams(str string, tempParams interface{}) (string, error) {
	// {{ will trigger pattern {{}} of actions of text/template
	r, _ := regexp.Compile(`(passw(or)?d|user\w*):\s*"?.*{{.*"?`)
	subString := r.FindString(str)
	if len(subString) > 0 {
		if match, _ := regexp.MatchString(`:\s*{{.*}}\s*$`, subString); !match {
			log.Errorf("Please avoid \"{{\" and \"}}\" characters unless mandatory in %s", subString)
			return "", GetError("errCustom")
		}
	}
	tpl := template.Must(template.New("StringTemplateConvertWithValue").Funcs(templateFuncs).Parse(str))
	var b bytes.Buffer
	if err := tpl.Execute(&b, tempParams); err != nil {
		return "", err
	}
	return b.String(), nil
}

func FileTemplateConvert(srcFile string, destFile string) error {
	data, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return err
	}

	err = CheckHashForContent(data, srcFile, "")
	if err != nil {
		if err.Error() == ERRORCODE_CHECK_HASH_FAIL {
			log.Warnln("Failed to check file hash:", err)
		} else {
			return err
		}

	}

	tpl := template.Must(template.New(srcFile).Funcs(templateFuncs).Parse(string(data)))
	var b bytes.Buffer
	if err := tpl.Execute(&b, templateParams); err != nil {
		return err
	}
	// #nosec G306
	return ioutil.WriteFile(destFile, b.Bytes(), 0644)
}

func LoadYamlFileToJson(file string) ([]byte, error) {
	yml, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	err = CheckHashForContent(yml, file, "")
	if err != nil {
		if err.Error() == ERRORCODE_CHECK_HASH_FAIL {
			log.Warnln("Failed to check file hash:", err)
		} else {
			return nil, err
		}

	}

	tpl := template.Must(template.New(file).Funcs(templateFuncs).Parse(string(yml)))
	var b bytes.Buffer
	if err := tpl.Execute(&b, templateParams); err != nil {
		return nil, err
	}
	yml = b.Bytes()
	return yaml.YAMLToJSON(yml)
}

func SaveJsonToYamlFile(json []byte, file string) error {
	yml, err := yaml.JSONToYAML(json)
	if err != nil {
		return err
	}
	// #nosec G306
	return ioutil.WriteFile(file, yml, 0644)
}

func SchemaStructToYaml(v SchemaStruct) (string, error) {
	json, err := v.MarshalBinary()
	if err != nil {
		return "", err
	}
	yml, err := yaml.JSONToYAML(json)
	return string(yml), err
}

func LoadSchemaStructFromYaml(v SchemaStruct, yml string) error {
	tpl := template.Must(template.New("tmp").Funcs(templateFuncs).Parse(yml))
	var b bytes.Buffer
	if err := tpl.Execute(&b, templateParams); err != nil {
		return err
	}
	json, err := yaml.YAMLToJSON(b.Bytes())
	if err != nil {
		return err
	}
	err = v.UnmarshalBinary(json)
	if err != nil {
		return err
	}
	return v.Validate(nil)
}

func LoadSchemaStructFromYamlFile(v SchemaStruct, file string) error {
	json, err := LoadYamlFileToJson(file)
	if err != nil {
		return err
	}
	err = v.UnmarshalBinary(json)
	if err != nil {
		return err
	}
	return v.Validate(nil)
}

func SaveSchemaStructToYamlFile(v SchemaStruct, file string) error {
	json, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return SaveJsonToYamlFile(json, file)
}

func ConvertSchemaStruct(from interface{}, to interface{}) error {
	json, err := json.Marshal(from)
	if err != nil {
		return err
	}
	err = to.(SchemaStruct).UnmarshalBinary(json)
	if err != nil {
		return err
	}
	err = to.(SchemaStruct).Validate(nil)
	if err != nil {
		return err
	}
	return nil
}

func LoadMapFromYamlFile(filePath string) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Errorf("Read file failed: %v ", err)
		return nil, GetError("errInvalidFile")
	}
	err = yaml.Unmarshal(b, &m)
	if err != nil {
		log.Errorf("Unmarshal failed: %v", err)
		return nil, GetError("errUnmarshal")
	}
	return m, nil
}

func ConvertStructToMap(v interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b, err := yaml.Marshal(v)
	if err != nil {
		log.Errorf("Marshal failed: %v", err)
		return nil, GetError("errMarshal")
	}
	if err = yaml.Unmarshal(b, &m); err != nil {
		log.Errorf("Unmarshal failed: %v", err)
		return nil, GetError("errUnmarshal")
	}
	return m, nil
}

func MergeMaps(a, b map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
