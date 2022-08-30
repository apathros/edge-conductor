/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

//go:generate mockgen -destination=./mock/fileutil_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt github.com/intel/edge-conductor/pkg/eputils FileWrapper

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Filelink string

type FileWrapper interface {
	IsValidFile(filename string) bool
	CheckFileLink(filename string) (Filelink, string)
	MakeDir(path string) error
	FileExists(filename string) bool
	RemoveFile(name string) error
	CopyFile(dstName, srcName string) (written int64, err error)
	WriteStringToFile(content, filename string) error
	DownloadFile(filepath string, fileurl string) error
	LoadJsonFromFile(filepath string, p interface{}) error
	CreateFolderIfNotExist(path string) error
	UncompressTgz(srctarfile, targetfolder string) error
	CompressTar(path string, tarFilePath string, perm os.FileMode) error
	GzipCompress(srctarfile, destpath string) error
}

const (
	Symbollink Filelink = "symbol"
	Hardlink   Filelink = "hard"
	Normalfile Filelink = "normal"
	Wrongfile  Filelink = ""
)

func GzipCompress(srctarfile, destpath string) error {
	readfile, err := os.Open(srctarfile)
	if err != nil {
		log.Errorf("Invalid Tar File")
		return err
	}
	srcfile := filepath.Base(srctarfile)
	destpath = filepath.Join(destpath, fmt.Sprintf("%s.gz", srcfile))
	if FileExists(destpath) {
		log.Infof("File already exist")
		return GetError("errFileExist")
	}
	gzfile, err := os.Create(destpath)
	if err != nil {
		log.Errorf("Failed to create gz file")
		return err
	}
	defer gzfile.Close()
	gztarfile := gzip.NewWriter(gzfile)
	gztarfile.Name = srcfile
	defer gztarfile.Close()

	_, err = io.Copy(gztarfile, readfile)
	return err
}

func IsValidFile(filename string) bool {
	flinktype, _ := CheckFileLink(filename)
	if flinktype == Normalfile || flinktype == Symbollink {
		return true
	}
	return false
}

func CheckFileLink(filename string) (Filelink, string) {
	fi, err := os.Lstat(filename)
	if err != nil {
		log.Errorf("Check link failure: %v", err)
		return Wrongfile, filename
	}
	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		log.Errorf("Check stat value failure: %v", err)
		return Wrongfile, filename
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		link, err := os.Readlink(filename)
		if err != nil {
			log.Errorf("Read symbol link error : %v", err)
			return Wrongfile, filename
		}
		return Symbollink, link
	}
	nlink := uint32(s.Nlink)
	if nlink > 1 {
		return Hardlink, filename
	} else {
		return Normalfile, filename
	}

}

func MakeDir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		log.Errorln("Cannot create the directory", path)
		return err
	}
	return nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func IsDirectory(filepath string) bool {
	pathinfo, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return false
	} else {
		return pathinfo.IsDir()
	}
}

func RemoveFile(name string) error {
	err := os.Remove(name)
	if err != nil && !os.IsNotExist(err) {
		log.Errorln("Failed to remove", name)
		return err
	}
	return nil
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	validfile := IsValidFile(srcName)
	log.Debugln("src file is:", srcName)
	if !validfile {
		return 0, GetError("errInvalidFile")
	}
	if FileExists(dstName) {
		validfile = IsValidFile(dstName)
		if !validfile {
			return 0, GetError("errInvalidFile")
		}
	}

	info, err := os.Stat(srcName)
	if err != nil {
		return 0, err
	}

	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func WriteStringToFile(content, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		log.Errorln("Failed to create", filename)
		return err
	}
	defer f.Close()

	file, err := os.Stat(filename)
	if err != nil {
		log.Errorln("Failed to stat", filename)
		return err
	}

	filemode_original := file.Mode()

	// Change file mode to 0600 to avoid modification
	// by other users during writing
	err = os.Chmod(filename, 0600)
	if err != nil {
		log.Errorln("Failed to change file mode: ", filename)
		return err
	}

	_, err = f.WriteString(content)
	if err != nil {
		log.Errorln("Failed to write", filename)
		return err
	}

	// Change file mode back
	err = os.Chmod(filename, filemode_original)
	if err != nil {
		log.Errorln("Failed to change file mode back: ", filename)
		return err
	}

	return nil
}

// Download file through url with go
func DownloadFile(filepath string, fileurl string) error {
	log.Infoln("Downloading:", fileurl, "to", filepath)
	ufile, _ := url.Parse(fileurl)
	if ufile.Scheme == "http" || ufile.Scheme == "https" {
		// Get the data
		// #nosec G107
		resp, err := http.Get(fileurl)
		if err != nil {
			log.Errorln("Failed to get", fileurl)
			return err
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			log.Errorln("Failed to create", filepath)
			return err
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		return err
	} else if ufile.Scheme == "file" {
		_, err := CopyFile(filepath, ufile.Path)
		return err
	} else {
		return GetError("errUrlSchema")
	}
}

func LoadJsonFromFile(filepath string, p interface{}) error {
	byteValue, err := LoadJsonFile(filepath)
	if err != nil {
		return err
	}
	err = json.Unmarshal(byteValue, &p)
	if err != nil {
		return err
	}

	return nil
}

func CreateFolderIfNotExist(path string) error {
	if !FileExists(path) {
		if err := MakeDir(path); err != nil {
			return err
		}
	}
	return nil
}

// UncompressTgz: extract source tarball to target folder.
//
// Parameters:
//   srctarfile:     source tarball file.
//   targetfolder:   target folder path.
//
func UncompressTgz(srctarfile, targetfolder string) error {
	if !FileExists(targetfolder) {
		if err := os.Mkdir(targetfolder, 0700); err != nil {
			log.Errorln("UncompressTgz: failed to create folder", targetfolder, err)
			return err
		}
	}

	f, err := os.OpenFile(srctarfile, os.O_RDONLY, 0600)
	if err != nil {
		log.Errorln("UncompressTgz: failed to open file", srctarfile, err)
		return err
	}
	defer f.Close()

	if !IsValidFile(srctarfile) {
		return GetError("errInvalidFile")
	}

	tarballData, err := gzip.NewReader(f)
	if err != nil {
		log.Errorln("UncompressTgz: failed to generate gzip reader", err)
		return err
	}

	reader := tar.NewReader(tarballData)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Errorln("UncompressTgz: failed to run reader.Next()", err)
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			folder := filepath.Join(targetfolder, header.Name)
			if err := os.Mkdir(folder, 0700); err != nil {
				log.Errorln("UncompressTgz: failed to create folder", folder, err)
				return err
			}
		case tar.TypeReg:
			filename := filepath.Join(targetfolder, header.Name)
			if FileExists(filename) {
				valid := IsValidFile(filename)
				if !valid {
					return GetError("errTgzUncompress")
				}
			}
			file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, os.FileMode(header.Mode))
			if err != nil {
				log.Errorln("UncompressTgz: failed to open file", file, err)
				return err
			}
			defer file.Close()
			// #nosec G110
			if _, err := io.Copy(file, reader); err != nil {
				log.Errorln("UncompressTgz: failed to copy data from reader to file", err)
				return err
			}
		case tar.TypeXHeader:
			log.Infof("header x: %v", header.PAXRecords)
		case tar.TypeXGlobalHeader:
			log.Infof("header g: %v", header.PAXRecords)
		default:
			log.Errorf("UncompressTgz: unknown header type %d in %s", header.Typeflag, header.Name)
			return GetError("errTgzHeader")
		}
	}

	return nil
}

// CompressTar: tar to path to a .tar file
//
// Parameters:
//   path:         a file path or a directory path to tar.
//   tarFilePath:  target tar file path.
//
func CompressTar(path string, tarFilePath string, perm os.FileMode) error {

	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	if strings.HasPrefix(tarFilePath, path) {
		log.Errorf("The target tar file path is in the path directory. path=%s, target file path=%s", path, tarFilePath)
		return GetError("errTarPath")
	}

	tarfile, err := os.OpenFile(tarFilePath, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		log.Errorf("CompressTar: create fail %s", tarFilePath)
		return err
	}
	defer tarfile.Close()

	valid := IsValidFile(tarFilePath)
	if !valid {
		return GetError("errInvalidFile")
	}

	tarball := tar.NewWriter(tarfile)
	defer tarball.Close()

	baseDir := ""
	if info.IsDir() {
		baseDir = filepath.Base(path)
	}

	return filepath.Walk(path,
		func(subPath string, info os.FileInfo, err error) error {
			if err != nil {
				log.Error(err)
				return err
			}
			header, err := tar.FileInfoHeader(info, info.Name())
			if err != nil {
				log.Errorf("CompressTar: fail to get %s header info, %s", subPath, err.Error())
				return err
			}
			if baseDir != "" {
				header.Name = filepath.Join(baseDir, strings.TrimPrefix(subPath, path))
			}

			if err := tarball.WriteHeader(header); err != nil {
				log.Errorf("CompressTar: fail to write %s header fail, %s", subPath, err.Error())
				return err
			}

			if info.IsDir() {
				return nil
			}

			file, err := os.OpenFile(subPath, os.O_RDONLY, 0600)
			if err != nil {
				log.Errorf("CompressTar: open %s fail, %s", subPath, err.Error())
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarball, file)
			return err
		})
}
