/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package eputils

import (
	papi "ep/pkg/api/plugins"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	configfolder   = "config/"
	confighashfile = "config/hash.yml"
)

var ConfigHashTable = &papi.Files{}

func init() {
	if FileExists(configfolder) {
		if err := loadHashCode(confighashfile); err != nil {
			log.Errorf("Read file %s err: %s", confighashfile, err)
			return
		}
	}
}

func loadHashCode(hashfile string) error {
	err := LoadSchemaStructFromYamlFile(ConfigHashTable, hashfile)
	if err != nil {
		log.Errorf("Read file %s err: %s", hashfile, err)
		return err
	}
	return nil
}

func getFileRealPath(filename string, workspace string) (string, error) {
	var fn string
	cwd := workspace

	if len(cwd) == 0 {
		if workingdir, err := os.Getwd(); err != nil {
			log.Errorln("Failed to get working dir:", err)
			return "", err
		} else {
			cwd = workingdir
		}
	}

	if strings.HasPrefix(filename, "workflow") {
		fn = filepath.Join(cwd, filename)
	} else {
		fn = filepath.Join(cwd, "config", filename)
	}

	return fn, nil
}

// Check hash for content
func CheckHashForContent(content []byte, file string, workspace string) error {
	var target string

	shortname, err := filepath.Rel("config/", file)
	if err != nil {
		target = file
	} else {
		target = shortname
	}

	for _, conf := range ConfigHashTable.Files {
		if conf.Hashtype != "sha256" {
			continue
		}
		filename := conf.Filename
		if filename != target {
			continue
		} else {
			filerealpath, err := getFileRealPath(target, workspace)
			if err != nil {
				return err
			}

			if err := CheckContentSHA256(content, conf.Hash); err != nil {
				log.Warnln("WARNING: Hash check failed for filename: " + filerealpath)
				log.Warnln("WARNING: The default configuration has been modified. User shall be responsible for the security and proper use of this tool with the configuration changes.")
				return GetError("errHash")
			}
		}

	}
	return nil
}

// CheckHash: Check hash code of config files under workspace folder.
func CheckHash(workspace string) error {

	for _, conf := range ConfigHashTable.Files {
		if conf.Hashtype != "sha256" {
			continue
		}

		target := conf.Filename
		filerealpath, err := getFileRealPath(target, workspace)
		if err != nil {
			return err
		}

		if err := CheckFileSHA256(filerealpath, conf.Hash); err != nil {
			log.Warnln("WARNING: Hash check failed.")
			log.Warnln("WARNING: The default configuration has been modified. User shall be responsible for the security and proper use of this tool with the configuration changes.")
			return GetError("errHash")
		}
	}
	return nil
}
