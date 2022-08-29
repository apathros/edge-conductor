/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

import (
	"bytes"
	pluginapi "ep/pkg/api/plugins"
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
)

const (
	sudoPrefix = "sudo "
)

//go:generate mockgen -destination=./mock/sshutil_mock/sshutil_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt ep/pkg/eputils SSHApiInterface,SSHDialInterface
type (
	SSHApiInterface interface {
		GenSSHConfig(server *pluginapi.Node) (*ssh.ClientConfig, error)
		RunRemoteCMD(addr string, cfg *ssh.ClientConfig, cmd string) error
		RunRemoteMultiCMD(addr string, cfg *ssh.ClientConfig, commands []string) error
		WriteRemoteFile(addr string, cfg *ssh.ClientConfig, content, path string) error
		CopyLocalFileToRemoteRootFileSudoNoPasswd(addr string, cfg *ssh.ClientConfig, localPath, remotePath string) error
		CopyLocalFileToRemoteFile(addr string, cfg *ssh.ClientConfig, localPath, remotePath string) error
		CopyRemoteRootFileToLocalFileSudoNoPasswd(addr string, cfg *ssh.ClientConfig, remotePath, localPath string, perm os.FileMode) error
		CopyRemoteFileToLocalFile(addr string, cfg *ssh.ClientConfig, remotePath, localPath string, perm os.FileMode) error
		ContainerdCertificatePathCreateSudoNoPasswd(addr string, cfg *ssh.ClientConfig, containerdcertpath, registry string) error
		ServiceRestartSudoNoPasswd(addr string, cfg *ssh.ClientConfig, serviceName string) error
		RemoteFileExists(addr string, cfg *ssh.ClientConfig, remotePath string) (bool, error)
		RunRemoteNodeMultiCMD(server *pluginapi.Node, commands []string) error
	}

	SSHDialInterface interface {
		Dial(network, addr string, config *ssh.ClientConfig) (*ssh.Client, error)
	}
)

var errRemoteNotAFile = errors.New("CopyRemoteFileToLocalFile: not a file")

func GenSSHConfig(server *pluginapi.Node) (*ssh.ClientConfig, error) {
	var config *ssh.ClientConfig
	if server.SSHKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(server.SSHKey))
		if err != nil {
			log.Errorf("Parse Key err: %v", err)
			return nil, err
		}
		config = &ssh.ClientConfig{
			Config: ssh.Config{
				Ciphers: []string{
					"aes256-gcm@openssh.com",
					"chacha20-poly1305@openssh.com",
					"aes256-ctr", "aes256-cbc"},
			},
			User: server.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			// #nosec G106
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	} else {
		config = &ssh.ClientConfig{
			Config: ssh.Config{
				Ciphers: []string{
					"aes256-gcm@openssh.com",
					"chacha20-poly1305@openssh.com",
					"aes256-ctr", "aes256-cbc"},
			},
			User: server.User,
			Auth: []ssh.AuthMethod{
				ssh.Password(server.SSHPasswd),
			},
			// #nosec G106
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}
	return config, nil
}

func RunRemoteCMD(addr string, cfg *ssh.ClientConfig, cmd string) error {
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		log.Errorf("Failed to create session: %v", err)
		return err
	}
	defer session.Close()
	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(cmd); err != nil {
		log.Errorf("Failed to run %s: %v ", cmd, err)
		return err
	}
	return nil
}

func RunRemoteMultiCMD(addr string, cfg *ssh.ClientConfig, commands []string) error {
	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Errorf("Failed to create session: %v", err)
		return err
	}
	defer session.Close()

	stdin, err := session.StdinPipe()
	if err != nil {
		log.Errorf("Failed to create session stdin: %v", err)
		return err
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr

	// Start remote shell
	err = session.Shell()
	if err != nil {
		log.Errorf("Failed to create session Shell: %v", err)
		return err
	}

	// append exit command for exit shell
	commands = append(commands, "exit")
	for _, cmd := range commands {
		_, err = fmt.Fprintf(stdin, "%s\n", cmd)
		if err != nil {
			log.Errorf("Failed to Run cmd %s: %v", cmd, err)
			return err
		}
	}

	// Wait for sess to finish
	err = session.Wait()
	if err != nil {
		log.Errorf("Failed to Wait Shell: %v", err)
		return err
	}
	return nil
}

func RunRemoteNodeMultiCMD(server *pluginapi.Node, commands []string) error {
	if server == nil {
		log.Warn("Can not get server info as it is empty.")
		return nil
	}

	cfg, err := GenSSHConfig(server)
	if err != nil {
		log.Errorf("Fail to gen config %v", err)
		return err
	}

	addr := fmt.Sprintf("%s:%d", server.IP, server.SSHPort)
	err = RunRemoteMultiCMD(addr, cfg, commands)
	log.Infof("sudo ssh command used!")
	if err != nil {
		return err
	}
	return nil
}

func WriteRemoteFile(addr string, cfg *ssh.ClientConfig, content, path string) error {

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}
	defer client.Close()
	sftpclient, err := sftp.NewClient(client)
	if err != nil {
		log.Errorf("Unable sftp to connect: %v", err)
		return err
	}
	err = sftpclient.MkdirAll(filepath.Dir(path))
	if err != nil {
		log.Errorf("Fail to mkdir %s, %v", path, err)
		return err
	}

	f, err := sftpclient.Create(path)
	if err != nil {
		log.Errorf("Fail to create %s, %v", path, err)
		return err
	}

	defer f.Close()

	err = f.Chmod(0600)
	if err != nil {
		log.Errorf("Fail to chmod %s, %v", path, err)
		return err
	}

	if _, err := f.Write([]byte(content)); err != nil {
		log.Errorf("Fail to write %s, %v", path, err)
		return err
	}

	return nil
}

func CopyLocalFileToRemoteRootFileSudoNoPasswd(addr string, cfg *ssh.ClientConfig, localPath, remotePath string) error {
	_, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}

	fileNameTemp := filepath.Base(localPath)
	remotePathTemp := "/tmp/" + fileNameTemp
	err = CopyLocalFileToRemoteFile(addr, cfg, localPath, remotePathTemp)
	if err != nil {
		log.Errorf("Failed to scp %s", localPath)
		return err
	}

	moveCmd := sudoPrefix + "mv " + remotePathTemp + " " + remotePath
	cmd := moveCmd
	err = RunRemoteCMD(addr, cfg, cmd)
	log.Infof("sudo ssh command used!")
	if err != nil {
		log.Errorf("Failed to scp %s", remotePath)
		return err
	}

	return nil
}

func CopyLocalFileToRemoteFile(addr string, cfg *ssh.ClientConfig, localPath, remotePath string) error {

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}
	defer client.Close()

	sftpclient, err := sftp.NewClient(client)
	if err != nil {
		log.Errorf("Unable sftp to connect: %v", err)
		return err
	}

	err = sftpclient.MkdirAll(filepath.Dir(remotePath))
	if err != nil {
		log.Errorf("Fail to mkdir %s, %v", remotePath, err)
		return err
	}

	dst, err := sftpclient.Create(remotePath)
	if err != nil {
		log.Errorf("Fail to create %s, %v", remotePath, err)
		return err
	}
	defer dst.Close()

	src, err := os.OpenFile(localPath, os.O_RDONLY, 0600)
	if err != nil {
		log.Errorf("CopyLocalFileToRemoteFile: open %s fail, %s", localPath, err.Error())
		return err
	}
	defer src.Close()

	err = dst.Chmod(0600)
	if err != nil {
		log.Errorf("CopyLocalFileToRemoteFile: Fail to chmod %s, %v", remotePath, err)
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		log.Errorf("CopyLocalFileToRemoteFile: copy fail %s", err.Error())
	}

	return nil
}

func CopyRemoteRootFileToLocalFileSudoNoPasswd(addr string, cfg *ssh.ClientConfig, remotePath, localPath string, perm os.FileMode) error {
	_, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}

	fileNameTemp := filepath.Base(remotePath)
	remotePathTemp := "/tmp/" + fileNameTemp

	cpCmd := sudoPrefix + "cp " + remotePath + " " + remotePathTemp
	chownCmd := sudoPrefix + "chown $(id -u " + cfg.User + "):$(id -g " + cfg.User + ")" + " " + remotePathTemp

	cmd := cpCmd + " && " + chownCmd
	err = RunRemoteCMD(addr, cfg, cmd)
	log.Infof("sudo ssh command used!")
	if err != nil {
		log.Errorf("Failed to copy file %v ", err)
		return err
	}

	err = CopyRemoteFileToLocalFile(addr, cfg, remotePathTemp, localPath, perm)
	if err != nil {
		log.Errorf("Failed to scp %s", remotePathTemp)
		return err
	}

	rmCmd := sudoPrefix + "rm " + " " + remotePathTemp
	cmd = rmCmd
	err = RunRemoteCMD(addr, cfg, cmd)
	log.Infof("sudo ssh command used!")
	if err != nil {
		log.Errorf("Failed to remove temp file %v ", err)
		return err
	}

	return nil
}

func CopyRemoteFileToLocalFile(addr string, cfg *ssh.ClientConfig, remotePath, localPath string, perm os.FileMode) error {

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}
	defer client.Close()

	sftpclient, err := sftp.NewClient(client)
	if err != nil {
		log.Errorf("Unable sftp to connect: %v", err)
		return err
	}

	info, err := sftpclient.Stat(remotePath)
	if err != nil {
		log.Errorf("Fail to access %s, %v", remotePath, err)
		return err
	}

	if info.IsDir() {
		log.Errorf("Must copy remote file not directory")
		return errRemoteNotAFile
	}

	src, err := sftpclient.Open(remotePath)
	if err != nil {
		log.Errorf("")
		return err
	}
	defer src.Close()

	err = os.MkdirAll(filepath.Dir(localPath), 0700)
	if err != nil {
		log.Errorf("Fail to mkdir %s, %v", localPath, err)
		return err
	}

	dst, err := os.OpenFile(localPath, os.O_RDWR|os.O_CREATE, perm)
	if err != nil {
		log.Errorf("CopyRemoteFileToLocalFile: create fail %s", localPath)
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		log.Errorf("CopyRemoteFileToLocalFile: copy fail %v", err.Error())
	}

	return nil
}

func ContainerdCertificatePathCreateSudoNoPasswd(addr string, cfg *ssh.ClientConfig, containerdcertpath, registry string) error {

	_, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}

	mkdirCertsdCmd := sudoPrefix + "mkdir -p " + containerdcertpath
	mkdirCertsdRegistryCmd := mkdirCertsdCmd + registry
	chownCmd := sudoPrefix + "chown $(id -u " + cfg.User + "):$(id -g " + cfg.User + ")" + " " + "-R" + " " + containerdcertpath
	cmd := mkdirCertsdCmd + " && " + mkdirCertsdRegistryCmd + " && " + chownCmd
	err = RunRemoteCMD(addr, cfg, cmd)
	log.Infof("sudo ssh command used!")
	if err != nil {
		log.Errorf("Failed to create CA root certificate path %v ", err)
		return err
	}

	return nil
}

func ServiceRestartSudoNoPasswd(addr string, cfg *ssh.ClientConfig, serviceName string) error {
	_, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return err
	}
	ServiceRestartCmd := sudoPrefix + "systemctl daemon-reload" + " && " + sudoPrefix + "systemctl restart " + serviceName
	cmd := ServiceRestartCmd
	err = RunRemoteCMD(addr, cfg, cmd)
	log.Infof("sudo ssh command used!")
	if err != nil {
		log.Errorf("Failed to restart service %v ", err)
		return err
	}

	return nil
}

func RemoteFileExists(addr string, cfg *ssh.ClientConfig, remotePath string) (bool, error) {

	client, err := ssh.Dial("tcp", addr, cfg)
	if err != nil {
		log.Errorf("Unable to connect:%s %v", addr, err)
		return false, err
	}
	defer client.Close()

	sftpclient, err := sftp.NewClient(client)
	if err != nil {
		log.Errorf("Unable sftp to connect: %v", err)
		return false, err
	}
	defer sftpclient.Close()

	remoteFileinfo, err := sftpclient.Stat(remotePath)
	log.Debugf("RemoteFileExists_remote path is %s, err is %v", remotePath, err)
	if os.IsNotExist(err) {
		return false, nil
	}
	if remoteFileinfo.IsDir() {
		log.Debugf("remote root CA path is a directory")
		return false, errRemoteNotAFile
	}

	return true, nil
}
