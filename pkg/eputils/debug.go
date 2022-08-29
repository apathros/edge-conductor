/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package eputils

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"path"
	"reflect"
	"runtime"
	"sigs.k8s.io/yaml"
)

func PPrint(v interface{}) {
	log.Infoln(reflect.TypeOf(v))
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		log.Infoln(fmt.Sprintf("\n%s\n", string(b)))
	}
}

func D(args ...interface{}) {
	s1 := fmt.Sprint(args...)
	_, file, line, _ := runtime.Caller(1)
	s2 := fmt.Sprintf("%s:%d ", path.Base(file), line)
	log.Debugf("%s%s", s2, s1)
}

func DumpVar(v interface{}) {
	yml, err := yaml.Marshal(v)
	if err == nil {
		log.Debugln(fmt.Sprintf("\n%s\n", string(yml)))
	} else {
		log.Debugln(err)
	}
}
