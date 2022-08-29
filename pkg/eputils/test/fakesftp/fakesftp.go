/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package fakesftp

import (
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//go:generate mockgen -destination=./mock/fakesftp_mock.go -package=mock -copyright_file=../../../../api/schemas/license-header.txt ep/pkg/eputils/test/fakesftp SftpClientInterface

type (
	SftpClientInterface interface {
		NewClient(conn *ssh.Client, opts ...sftp.ClientOption) (*sftp.Client, error)
	}
)
