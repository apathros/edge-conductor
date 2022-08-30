/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"errors"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	troubleshootfolder = "docs/troubleshooting"
	indexfilename      = "index.md"
	ecfile             = "pkg/eputils/errorcode.go"
)

const header = `
[Edge Conductor]: https://github.com/intel/edge-conductor
[Troubleshooting]: ./index.md
[Edge Conductor] / [Troubleshooting]
# Edge Conductor Troubleshooting
This document provides more detailed information about the error code you have encountered during the build/deploy cluster or service on Edge Conductor.
`

const footer = `
Copyright (C) 2022 Intel Corporation
SPDX-License-Identifier: Apache-2.0
`

type ecSubGroup struct {
	ecGroup      map[string]string
	ecComment    string
	ecSubComment []string
}

/*
break all errorGroup into subGroup depends on their ecode
ie. E001.*** belongs to E001 ecSubGroup
*/
func setupEcGroups() (map[string]ecSubGroup, error) {
	ecMap := make(map[string]ecSubGroup)
	for _, v := range eputils.ErrorGroup {
		errContent := v.(*eputils.EC_errors)
		code := errContent.Code()
		msg := errContent.Msg()
		ecSubGroupIndex := strings.Split(code, ".")
		if len(ecSubGroupIndex) != 2 {
			return nil, errors.New(code + " has syntax error")
		}
		_, err := strconv.Atoi(ecSubGroupIndex[1])
		if err != nil {
			return nil, errors.New(code + " has syntax error")
		}
		if !strings.HasPrefix(ecSubGroupIndex[0], "E") {
			return nil, errors.New(code + " has syntax error")
		}
		if _, ok := ecMap[ecSubGroupIndex[0]]; !ok {
			subGroup := new(ecSubGroup)
			errcodeMap := make(map[string]string)
			errcodeMap[ecSubGroupIndex[1]] = msg
			subGroup.ecGroup = errcodeMap
			ecMap[ecSubGroupIndex[0]] = *subGroup
		} else {
			subGroup := ecMap[ecSubGroupIndex[0]]
			subGroup.ecGroup[ecSubGroupIndex[1]] = msg
		}

	}
	return ecMap, nil
}

/*
add description to each ecSubGroup from comments in errorcode.go file
*/
func setupEcGroupComments(ecMap map[string]ecSubGroup) error {
	content, err := ioutil.ReadFile(ecfile)
	if err != nil {
		return err
	}

	for k, _ := range ecMap {
		rComment, _ := regexp.Compile(`\s*` + k + `\s*:.+`)
		rSubComment, _ := regexp.Compile(`//\s*` + k + `\.[0-9]+.*:.+`)
		comment := rComment.FindString(string(content))
		subComment := rSubComment.FindAllString(string(content), -1)
		if len(comment) > 0 {
			if entry, ok := ecMap[k]; ok {
				entry.ecComment = comment
				ecMap[k] = entry
			}
			if len(subComment) > 0 {
				if entry, ok := ecMap[k]; ok {
					entry.ecSubComment = subComment
					ecMap[k] = entry
				}
			}
		}
	}
	return nil
}

func genEcGroupContent(ecMap map[string]ecSubGroup) string {
	var content string
	keys := make([]string, 0, len(ecMap))

	for k := range ecMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, key := range keys {
		ecSubGroupKey := make([]string, 0, len(ecMap[key].ecGroup))

		for k := range ecMap[key].ecGroup {
			ecSubGroupKey = append(ecSubGroupKey, k)
		}

		sort.Strings(ecSubGroupKey)
		content += "## " + ecMap[key].ecComment + "\n"

		for _, k := range ecSubGroupKey {
			ecode := key + "." + k
			if len(ecMap[key].ecSubComment) > 0 {
				for i, s := range ecMap[key].ecSubComment {
					if match, _ := regexp.MatchString(ecode[0:6], s); match {
						content += "\n" + s + "\n"
						ecSubComment := append(ecMap[key].ecSubComment[:i], ecMap[key].ecSubComment[i+1:]...)
						if entry, ok := ecMap[key]; ok {
							entry.ecSubComment = ecSubComment
							ecMap[key] = entry
						}
					}
				}
			}

			linkname := strings.ReplaceAll(ecode, "E", "e") + ".md"
			linkfile := filepath.Join(troubleshootfolder, linkname)
			if eputils.FileExists(linkfile) {
				content += "* [" + ecode + "](./" + linkname + "): " + ecMap[key].ecGroup[k] + "\n"
			} else {
				content += "* " + ecode + ": " + ecMap[key].ecGroup[k] + "\n"
			}
		}
	}
	return content
}

func genDocMd(ecMap map[string]ecSubGroup) {
	if ecMap == nil {
		return
	}
	indexfile := filepath.Join(troubleshootfolder, indexfilename)
	if !eputils.IsDirectory(troubleshootfolder) {
		eputils.MakeDir(troubleshootfolder)
	} else if eputils.FileExists(indexfile) {
		eputils.RemoveFile(indexfile)
	}
	var content string
	content += header + "\n"
	content += genEcGroupContent(ecMap)
	content += footer

	eputils.WriteStringToFile(content, indexfile)
}

func main() {
	ecMap, _ := setupEcGroups()
	err := setupEcGroupComments(ecMap)
	if err != nil {
		log.Fatal(err)
	}
	genDocMd(ecMap)
}
