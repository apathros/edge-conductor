/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

// CheckContentSHA256 checks the content
func CheckContentSHA256(content []byte, sha256expected string) error {
	hashhandler := sha256.New()
	hashhandler.Write(content)
	if fmt.Sprintf("%x", hashhandler.Sum(nil)) != sha256expected {
		log.Warnln("SHA256 check failed for content")
		return GetError("errShaCheckFailed")
	}
	return nil
}

// CheckFileSHA256: Check if a file's SHA256 Sum is expected.
//
// Parameters:
//   filename:        File to be checked.
//   sha256expected:  Expected SHA256 Sum.
//
func CheckFileSHA256(filename, sha256expected string) error {
	f, err := os.Open(filename)
	if err != nil {
		log.Errorln("Failed to read", filename, err)
		return err
	}
	defer f.Close()

	hashhandler := sha256.New()

	if _, err := io.Copy(hashhandler, f); err != nil {
		log.Errorln("Failed to read", filename, err)
		return err
	}

	if fmt.Sprintf("%x", hashhandler.Sum(nil)) != sha256expected {
		log.Warnln("SHA256 check failed for", filename)
		return GetError("errShaCheckFailed")
	}
	return nil
}

// CheckFileDescriptorSHA256 checks the content which the file descriptor points to
func CheckFileDescriptorSHA256(f *os.File, sha256expected string) error {
	hashhandler := sha256.New()
	if _, err := io.Copy(hashhandler, f); err != nil {
		log.Errorln("Failed to read file descriptor", err)
		return err
	}

	if fmt.Sprintf("%x", hashhandler.Sum(nil)) != sha256expected {
		log.Warnln("SHA256 check failed for file descriptor")
		return GetError("errShaCheckFailed")
	}
	return nil
}

func GenFileSHA256(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		log.Errorln("Failed to read", filename, err)
		return "", err
	}
	defer f.Close()

	hashhandler := sha256.New()
	if _, err := io.Copy(hashhandler, f); err != nil {
		log.Errorln("Failed to read", filename, err)
		return "", err
	}

	return fmt.Sprintf("%x", hashhandler.Sum(nil)), nil
}
