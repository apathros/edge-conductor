/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package orasutils

//go:generate mockgen -destination=./mock/orasutil_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt github.com/intel/edge-conductor/pkg/eputils/orasutils OrasInterface,OrasUtilInterface

import (
	sysctx "context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/deislabs/oras/pkg/content"
	"github.com/deislabs/oras/pkg/context"
	"github.com/deislabs/oras/pkg/oras"
	dockerutils "github.com/intel/edge-conductor/pkg/eputils/docker"

	ctndcontent "github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/docker/api/types"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/intel/edge-conductor/pkg/eputils"
)

type (
	OrasClient struct {
		hosts map[string]OrasHost
	}

	OrasHost struct {
		address     string
		resolverOpt docker.ResolverOptions
		resolver    remotes.Resolver
	}

	OrasInterface interface {
		Push(sysctx.Context, remotes.Resolver, string, ctndcontent.Provider, []ocispec.Descriptor, ...oras.PushOpt) (ocispec.Descriptor, error)
		Pull(sysctx.Context, remotes.Resolver, string, ctndcontent.Ingester, ...oras.PullOpt) (ocispec.Descriptor, []ocispec.Descriptor, error)
	}

	OrasUtilInterface interface {
		OrasPushFile(filename, subRef, rev string) (string, error)
		OrasPullFile(targetFile string, regRef string) error
	}
)

var (
	OrasCli *OrasClient
)

const (
	ConfigMediaType = "application/vnd.edge.conductor.config.v1+json"
	FileMediaType   = "application/edge.conductor.file"
	RegProject      = "library"
)

func (c *OrasClient) OrasPushFile(filename, subRef, rev string) (string, error) {
	if filename == "" {
		return "", eputils.GetError("errFileEmpty")
	}
	if subRef == "" {
		subRef = "tmp"
	}
	if rev == "" {
		rev = "0.0.0"
	}
	var targetRef string
	mediaType := FileMediaType
	ctx := context.Background()

	if h, exists := c.hosts["default"]; exists {
		err := updateResolver(h.address, c)
		if err != nil {
			return "", err
		}
		store := content.NewFileStore(".")
		defer store.Close()
		filebase := filepath.Base(filename)
		targetRef = fmt.Sprintf("%s/%s/%s/%s:%s", h.address, RegProject, subRef, filebase, rev)
		desc, err := store.Add(filebase, mediaType, filename)
		if err != nil {
			log.Errorf("Failed to open: %s", filename)
			return "", err
		}

		pushContents := []ocispec.Descriptor{desc}

		log.Infof("Push file %s to %s", filename, targetRef)
		_, err = oras.Push(ctx, h.resolver, targetRef, store, pushContents,
			oras.WithConfigMediaType(ConfigMediaType))
		if err != nil {
			return "", err
		}
	} else {
		log.Errorf("Oras default resolver %s not found", h.address)
		return "", eputils.GetError("errOrasDefaultResolver")
	}
	ref := fmt.Sprintf("oci://%s", targetRef)
	return ref, nil
}

func (c *OrasClient) OrasPullFile(targetFile string, regRef string) error {
	mediaType := FileMediaType
	ctx := context.Background()
	u, err := url.Parse(regRef)
	if err != nil {
		return err
	}
	targetRef := strings.TrimPrefix(regRef, "oci://")
	if h, exists := c.hosts[u.Host]; exists {
		store := content.NewFileStore(".")
		defer store.Close()
		allowedMediaTypes := []string{
			mediaType,
		}

		log.Infof("Pulling from %s", targetRef)
		_, _, err := oras.Pull(ctx, h.resolver, targetRef, store,
			oras.WithAllowedMediaTypes(allowedMediaTypes),
			// Remap the local path of the file to target path.
			oras.WithPullBaseHandler(images.HandlerFunc(
				func(ctx sysctx.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
					if desc.MediaType == mediaType {
						name, _ := content.ResolveName(desc)
						store.MapPath(name, targetFile)
					}
					return nil, nil
				})))
		if err != nil {
			log.Errorf("Error pulling from %s: %v", targetRef, err)
			return err
		}
	} else {
		log.Errorf("Oras resolver %s not found", u.Host)
		return eputils.GetError("errOrasResolver")
	}
	return nil
}

func OrasNewClient(authConf *types.AuthConfig, cacert string) error {
	if OrasCli == nil {
		OrasCli = &OrasClient{
			hosts: make(map[string]OrasHost),
		}
	}

	tlsCfg, err := dockerutils.GetCertPoolCfgWithCustomCa(cacert)
	if err != nil {
		return err
	}

	var httpClient *http.Client
	if tlsCfg == nil {
		httpClient = http.DefaultClient
	} else {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsCfg,
			},
		}
	}

	opts := docker.ResolverOptions{}
	opts.Client = httpClient
	opts.Credentials = func(hostName string) (string, string, error) {
		if hostName == authConf.ServerAddress {
			return authConf.Username, authConf.Password, nil
		} else {
			return "", "", nil
		}
	}
	if _, exists := OrasCli.hosts["default"]; !exists {
		OrasCli.hosts["default"] = OrasHost{
			address:     authConf.ServerAddress,
			resolverOpt: opts,
			resolver:    docker.NewResolver(opts),
		}
	}
	OrasCli.hosts[authConf.ServerAddress] = OrasHost{
		address:     authConf.ServerAddress,
		resolverOpt: opts,
		resolver:    docker.NewResolver(opts),
	}

	return nil
}

// This is a workaround for contianerd resovler authentication issue
// https://github.com/containerd/containerd/issues/4379
func updateResolver(host string, c *OrasClient) error {
	if c.hosts["default"].address == host {
		defaultOpt := c.hosts["default"].resolverOpt
		OrasCli.hosts["default"] = OrasHost{
			address:     host,
			resolverOpt: defaultOpt,
			resolver:    docker.NewResolver(defaultOpt),
		}
	}
	if _, exists := c.hosts[host]; exists {
		previousOpt := c.hosts[host].resolverOpt
		OrasCli.hosts[host] = OrasHost{
			address:     host,
			resolverOpt: previousOpt,
			resolver:    docker.NewResolver(previousOpt),
		}
	} else {
		log.Errorf("Oras resolver %s not found", host)
		return eputils.GetError("errOrasResolver")
	}
	return nil
}
