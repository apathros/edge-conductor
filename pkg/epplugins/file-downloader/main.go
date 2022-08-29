/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package filedownloader

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	eputils "ep/pkg/eputils"
	repoutils "ep/pkg/eputils/repoutils"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_files := input_files(in)

	output_files := output_files(outp)

	tmpFolder := filepath.Join(input_ep_params.Runtimedir, "tmp")
	defer func() {
		err := os.RemoveAll(tmpFolder)
		if err != nil {
			log.Errorln("failed to remove", tmpFolder, err)
		}
	}()

	for _, file := range input_files.Files {
		fileurl := file.URL

		fileName := path.Base(fileurl)

		targetFile := filepath.Join(tmpFolder, fileName)
		if err := eputils.CreateFolderIfNotExist(path.Dir(targetFile)); err != nil {
			return err
		}

		log.Infof("Downloading %s", fileurl)
		if err := eputils.DownloadFile(targetFile, fileurl); err != nil {
			log.Errorln("Failed to download", fileurl)
			log.Errorln(err)
			return eputils.GetError("errDownload")

		} else {
			// Need to check hash if provided.
			if len(file.Hash) > 0 {
				log.Infof("Checking file hash of type %s.", strings.ToUpper(file.Hashtype))
				if file.Hashtype == "sha256" {
					if err := eputils.CheckFileSHA256(targetFile, file.Hash); err != nil {
						log.Errorln("Download failed: SHA256 check failed for", fileurl)
						if err := eputils.RemoveFile(targetFile); err != nil {
							log.Errorln(err)
						}
						return err
					}
				}
				log.Infoln("File Hash Check successfully for", fileurl)
			}

			log.Infof("Downloaded successfully.")

			ref, err := repoutils.PushFileToRepo(targetFile, file.Urlreplacement.New, "")
			if err != nil {
				return err
			}

			if err := eputils.RemoveFile(targetFile); err != nil {
				log.Errorln(err)
				return err
			}
			file.Mirrorurl = ref
			output_files.Files = append(
				output_files.Files,
				file)
		}
	}

	return nil
}
