/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"os"
	"path/filepath"
	"text/template"

	log "github.com/sirupsen/logrus"

	eputils "ep/pkg/eputils"
)

const (
	configs_folder = "configs/"
	hash_file      = "_workspace/config/hash.yml"
)

type configStruct struct {
	Filename, HashString, HashType string
}

/* Define Templates */

// Template for hash.yml.
const template_hash_yml = `
# Auto generated, do not modify.
files:
{{- block "list" .}}{{range .}}
- filename:   "{{.Filename}}"
  hash: "{{.HashString}}"
  hashtype:   "{{.HashType}}"
{{- end -}}{{- end}}
`

func RemoveFile(name string) error {
	err := os.Remove(name)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func main() {
	tphashgo := template.Must(template.New("hashgo").Parse(template_hash_yml))

	var confs []configStruct

	// Get File List
	var files []string

	if err := filepath.Walk(configs_folder,
		func(fpath string, finfo os.FileInfo, err error) error {
			if err == nil && !finfo.IsDir() {
				files = append(files, fpath)
			}
			return err
		}); err != nil {
		log.Fatal(err)
	}

	for _, fn := range files {
		log.Infoln("Check hashcode for", fn)
		if hash, err := eputils.GenFileSHA256(fn); err != nil {
			log.Errorln("Failed to generate SHA256 hashcode.")
		} else {
			log.Infoln("  ", hash)
			if rfn, err := filepath.Rel(configs_folder, fn); err != nil {
				log.Errorln("Failed to get relative path:", err)
			} else {
				confs = append(confs,
					configStruct{
						Filename:   rfn,
						HashString: hash,
						HashType:   "sha256",
					})
			}
		}
	}

	// Generate hash.go
	err := RemoveFile(hash_file)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(hash_file)
	if err != nil {
		log.Fatal(err)
	}
	err = tphashgo.Execute(f, confs)
	if err != nil {
		log.Fatal(err)
	}
}
