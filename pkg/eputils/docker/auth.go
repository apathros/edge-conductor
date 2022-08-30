/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package docker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	registrytypes "github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/errdefs"
	"github.com/docker/docker/registry"
	log "github.com/sirupsen/logrus"
)

func GetAuthConf(server, port, user, password string) (*types.AuthConfig, error) {
	if server == "" || port == "" {
		return nil, eputils.GetError("errNoRegistry")
	}
	registryhead := fmt.Sprintf("%s:%s", server, port)

	auth := &types.AuthConfig{
		Username:      user,
		Password:      password,
		ServerAddress: registryhead,
	}

	return auth, nil
}

func GetCertPoolCfgWithCustomCa(caCertFile string) (*tls.Config, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if eputils.FileExists(caCertFile) {
		customCaCert, err := ioutil.ReadFile(caCertFile)
		if err != nil {
			log.Errorf("Failed to append %s to RootCAs: %v", caCertFile, err)
			return nil, err
		}

		rootCAs.AppendCertsFromPEM(customCaCert)
		tlscfg := &tls.Config{
			RootCAs: rootCAs,
		}
		return tlscfg, nil
	} else {
		log.Warnf("CA certificate is not provided. Use system CA cert.")
		return nil, nil
	}
}

func LoadDockerCliCredentials(imgName string) (*types.AuthConfig, error) {
	imageRef, err := reference.ParseNormalizedNamed(imgName)
	if err != nil {
		log.Warnf("failed to parse imgName %s with error %s", imgName, err)
		return nil, nil
	}

	var repoInfo *registry.RepositoryInfo
	repoInfo, err1 := registry.ParseRepositoryInfo(imageRef)
	if err1 != nil {
		log.Warnf("failed to parse RepositoryInfo with error %s", err1)
		return nil, nil
	}

	authConfig, err := getAuthFromCliConfig(repoInfo.Index)

	if err == nil {
		log.Debugln("Got auth from docker cli config, will download images from docker hub as an authenticated user.")
		return authConfig, nil
	} else {
		if errdefs.IsUnauthorized(err) {
			log.Debugln(err)
			log.Debugln("to download images from docker hub as an authenticated user, please docker login/relogin.")
		} else if errdefs.IsDataLoss(err) {
			log.Warnln(err)
			log.Warnln("to download images from docker hub as an authenticated user, please docker login/relogin.")
		} else {
			log.Warnf("failed to get auth from docker cli config for %s.", err)
		}
		return nil, nil
	}
}

func getAuthFromCliConfig(registryIndex *registrytypes.IndexInfo) (*types.AuthConfig, error) {
	cliAuthConfigs, err := getCliAuthConfig()

	if cliAuthConfigs != nil {
		authConfig := registry.ResolveAuthConfig(*cliAuthConfigs, registryIndex)
		if authConfig.Auth == "" {
			return nil, errdefs.Unauthorized(eputils.GetError("errNoAuth"))
		}

		err = parseAuth(&authConfig)
		if err == nil {
			return &authConfig, nil
		}

	}

	return nil, err
}

// the first return value should be nil,
// unless the auth configs map got from config file was not empty.
func getCliAuthConfig() (*map[string]types.AuthConfig, error) {
	configFilePath := getCliConfigFilePath()

	if configFilePath == "" {
		return nil, eputils.GetError("errInvalidPath")
	}

	log.Infof("Read docker cli config from %s", configFilePath)
	if !eputils.FileExists(configFilePath) {
		log.Errorf("docker client configuration file %s not found", configFilePath)
		return nil, errdefs.Unauthorized(eputils.GetError("errDockerCltCfg"))
	}

	if configCli, err := ioutil.ReadFile(configFilePath); err != nil {
		log.Errorf("Failed open docker client config file %s with error %s", configFilePath, err)
		return nil, eputils.GetError("errDockerClientConfig")
	} else {
		authConfig := map[string]types.AuthConfig{}
		err = parseCliConfigToAuthConfigMap(configCli, &authConfig)

		if err != nil {
			log.Errorf("Failed to parse cli config, because of failing to parse client configuration file to AuthConfig map with error %s", err)
			return nil, errdefs.DataLoss(eputils.GetError("errParseCliCfg"))
		} else {
			return &authConfig, nil
		}
	}
}

func parseAuth(authConfig *types.AuthConfig) error {
	parsedToken, e := base64.StdEncoding.DecodeString(authConfig.Auth)
	if e != nil {
		log.Errorf("Failed to decode auth from %s", authConfig.Auth)
		return errdefs.DataLoss(eputils.GetError("errDecode"))
	}

	parts := bytes.SplitN(parsedToken, []byte(":"), 2)
	if len(parts) == 2 {
		authConfig.Username = string(parts[0])
		authConfig.Password = string(parts[1])
		return nil
	} else {
		log.Errorf("invalid auth string: %s.", parsedToken)
		return errdefs.DataLoss(eputils.GetError("errInvalidString"))
	}
}

func parseCliConfigToAuthConfigMap(cliConfig []byte, authConfigMap *map[string]types.AuthConfig) error {
	var objmap map[string]json.RawMessage
	if err := json.Unmarshal(cliConfig, &objmap); err != nil {
		return err
	}

	if _, ok := objmap["auths"]; ok {
		if err := json.Unmarshal(objmap["auths"], authConfigMap); err != nil {
			return err
		}
	}
	return nil
}

func getCliConfigFilePath() string {
	if p := os.Getenv("DOCKER_CONFIG"); p != "" {
		return filepath.Join(p, "config.json")
	} else if p := os.Getenv("HOME"); p != "" {
		return filepath.Join(p, ".docker", "config.json")
	} else {
		return ""
	}
}
