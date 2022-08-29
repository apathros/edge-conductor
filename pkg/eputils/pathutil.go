/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//go:generate mockgen -destination=./mock/pathutil_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt ep/pkg/eputils PathWrapper

package eputils

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/url"
	"path"
)

type PathWrapper interface {
	GetBaseUrl(originalURL string) string
}

func GetBaseUrl(originalURL string) string {
	u, err := url.Parse(originalURL)
	if err != nil {
		log.Errorln(err)
		return ""
	}
	if u.User != nil {
		userinfo := u.User.Username()
		if p, yes := u.User.Password(); yes {
			userinfo = userinfo + ":" + p
		}
		return fmt.Sprintf("%s://%s@%s%s", u.Scheme, userinfo, u.Host, path.Dir(u.Path))
	} else {
		return fmt.Sprintf("%s://%s%s", u.Scheme, u.Host, path.Dir(u.Path))
	}
}
