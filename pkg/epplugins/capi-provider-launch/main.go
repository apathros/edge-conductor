/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capiproviderlaunch

import (
	//"bytes"
	"bytes"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	capiutils "github.com/intel/edge-conductor/pkg/eputils/capiutils"
	cutils "github.com/intel/edge-conductor/pkg/eputils/conductorutils"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	"os"
	"os/exec"
	"path"

	//	"path"
	"path/filepath"
	"strings"

	"text/template"

	log "github.com/sirupsen/logrus"
)

type ProviderConfig struct {
	Name, Version, Label string
}
type BinariesConfig struct {
	Name, Version, Url string
}
type CertManagerConfig struct {
	Version string
}

type ClusterCtlConfig struct {
	CoreProvider, BootstrapProvider, ControlPlaneProvider, InfrastructureProvider *ProviderConfig
	CertManager                                                                   *CertManagerConfig
	RuntimeDir                                                                    string
	Binaries                                                                      *BinariesConfig
}

const template_config_yaml = `
cert-manager:
  version: "{{ .CertManager.Version }}"
  url: "{{ .RuntimeDir }}/cert-manager/{{ .CertManager.Version }}/cert-manager.yaml"
providers:
- name: "{{ .CoreProvider.Name }}"
  type: "CoreProvider"
  url: "{{ .RuntimeDir }}/{{ .CoreProvider.Label}}/{{ .CoreProvider.Version}}/core-components.yaml"
- name: "{{ .BootstrapProvider.Name }}"
  type: "BootstrapProvider"
  url: "{{ .RuntimeDir }}/{{ .BootstrapProvider.Label}}/{{ .BootstrapProvider.Version}}/bootstrap-components.yaml"
- name: "{{ .ControlPlaneProvider.Name }}"
  type: "ControlPlaneProvider"
  url: "{{ .RuntimeDir }}/{{ .ControlPlaneProvider.Label}}/{{ .ControlPlaneProvider.Version}}/control-plane-components.yaml"
- name: "{{ .InfrastructureProvider.Name }}"
  type: "InfrastructureProvider"
  url: "{{ .RuntimeDir }}/{{ .InfrastructureProvider.Label}}/{{ .InfrastructureProvider.Version}}/infrastructure-components.yaml"
`

const kind_config_yaml = `
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: "{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}"
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 8088
    protocol: TCP
  - containerPort: 443
    hostPort: 9443
    protocol: TCP
  extraMounts:
    - containerPath: /etc/containerd/certs.d/
      hostPath: {{ .Runtimedir }}/data/cert/
containerdConfigPatches:
  - |-
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."k8s.gcr.io"]
        endpoint = ["https://{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}:{{ .Kitconfig.Parameters.GlobalSettings.RegistryPort }}/v2/k8s.gcr.io", "https://k8s.gcr.io"]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."gcr.io"]
        endpoint = ["https://{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}:{{ .Kitconfig.Parameters.GlobalSettings.RegistryPort }}/v2/gcr.io", "https://gcr.io"]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."quay.io"]
        endpoint = ["https://{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}:{{ .Kitconfig.Parameters.GlobalSettings.RegistryPort }}/v2/quay.io", "https://quay.io"]
      [plugins."io.containerd.grpc.v1.cri".registry.mirrors."projects.registry.vmware.com"]
        endpoint = ["https://{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}:{{ .Kitconfig.Parameters.GlobalSettings.RegistryPort }}/v2/projects.registry.vmware.com/", "https://projects.registry.vmware.com/"]
      [plugins."io.containerd.grpc.v1.cri".registry.configs."{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}:{{ .Kitconfig.Parameters.GlobalSettings.RegistryPort }}".tls]
        ca_file = "/etc/containerd/certs.d/{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}:{{ .Kitconfig.Parameters.GlobalSettings.RegistryPort }}/ca.crt"
      [plugins."io.containerd.grpc.v1.cri".registry.configs."{{ .Kitconfig.Parameters.GlobalSettings.ProviderIP }}:{{ .Kitconfig.Parameters.GlobalSettings.RegistryPort }}".auth]
        username = "{{ .Kitconfig.Parameters.Customconfig.Registry.User }}"
        password = "{{ .Kitconfig.Parameters.Customconfig.Registry.Password }}"
`

func launchManagementCluster(ep_params *pluginapi.EpParams, clusterManifest *pluginapi.Clustermanifest, files *pluginapi.Files) error {
	mgr_cluster_kubeconfig := ""
	var err error

	for _, ext := range ep_params.Extensions {
		if ext.Name == capiutils.CAPI_BYOH || ext.Name == capiutils.CAPI_METAL3 {
			for _, ext_section := range ext.Extension.Extension {
				if ext_section.Name == capiutils.EXTENSION_INFRA_PROVIDER {
					for _, config := range ext_section.Config {
						if config.Name == capiutils.CONFIG_MANAGEMENT_CLUSTER_KUBECONFIG {
							mgr_cluster_kubeconfig = config.Value
						}
					}
				}
			}
		}
	}

	if mgr_cluster_kubeconfig != "" {
		log.Infof("User provides management cluster")
		m_kubeconfig := filepath.Join(ep_params.Runtimedir, capiutils.MANAGEMENT_KUBECONFIG)
		_, err = eputils.CopyFile(m_kubeconfig, mgr_cluster_kubeconfig)
		if err != nil {
			log.Errorf("Failed to copy %s", mgr_cluster_kubeconfig)
			return err
		}
		return nil
	}

	kindURL := ""
	for _, file := range files.Files {
		if strings.Contains(file.Mirrorurl, "capi/kind") {
			kindURL = file.Mirrorurl
		}
	}

	if kindURL == "" {
		return eputils.GetError("errNoKindBinInReg")
	}

	kindBin := filepath.Join(ep_params.Runtimebin, "kind")
	err = repoutils.PullFileFromRepo(kindBin, kindURL)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}
	err = os.Chmod(kindBin, 0700)
	if err != nil {
		log.Errorf("%v", err)
		return err
	}

	kindTemplateConfigPath := filepath.Join(ep_params.Runtimedir, "kindTemplateconfig.yaml")
	err = eputils.WriteStringToFile(kind_config_yaml, kindTemplateConfigPath)
	if err != nil {
		log.Errorf("Write kind config template file fail, %v", err)
		return err
	}

	kindConfigPath := filepath.Join(ep_params.Runtimedir, "kindconfig.yaml")
	err = eputils.FileTemplateConvert(kindTemplateConfigPath, kindConfigPath)
	if err != nil {
		log.Errorf("Gent kind config fail, %v", err)
		return err
	}

	kubeconfigPath := filepath.Join(ep_params.Runtimedir, capiutils.MANAGEMENT_KUBECONFIG)
	if ep_params.Kitconfig == nil || ep_params.Kitconfig.Parameters == nil ||
		ep_params.Kitconfig.Parameters.GlobalSettings == nil {
		return eputils.GetError("errKitConfigParm")
	}
	kindprovider, err := cutils.GetClusterManifest(clusterManifest, "kind")
	if err != nil {
		return eputils.GetError("errKitConfigParm")
	}
	kindImageNode, err := cutils.GetImageFromProvider(kindprovider, "img_node")
	if err != nil {
		return eputils.GetError("errKitConfigParm")
	}

	imageUrl := ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP +
		":" + ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort +
		"/docker.io/" + kindImageNode

	// If mngr cluster exist, it will be deleled first.
	cmd := exec.Command(kindBin, "delete", "cluster", "--name", capiutils.MANAGEMENT_CLUSTER_NAME)
	_, err = eputils.RunCMD(cmd)
	if err != nil {
		log.Errorf("Failed to delete existed management cluster, %v", err)
		return eputils.GetError("errRunClusterctlCmd")
	}

	log.Infof("Starting management cluster ...")
	cmd = exec.Command(kindBin, "create", "cluster", "--name", capiutils.MANAGEMENT_CLUSTER_NAME, "--config", kindConfigPath, "--kubeconfig", kubeconfigPath, "--image", imageUrl)
	_, err = eputils.RunCMD(cmd)

	if err != nil {
		log.Errorf("Create management cluster fail, %v", err)
		return eputils.GetError("errRunClusterctlCmd")
	}

	os.Remove(kindTemplateConfigPath)
	os.Remove(kindConfigPath)

	return nil
}

func generateClusterCtlConfig(manifest []*pluginapi.ClustermanifestCapiClusterProvidersItems0, infra_provider capiutils.CapiInfraProvider, ep_params *pluginapi.EpParams) (config ClusterCtlConfig, err error) {
	config.RuntimeDir = filepath.Join(ep_params.Runtimedir, "clusterapi")

	config_name := capiutils.GetManifestConfigNameByCapiInfraProvider(infra_provider)

	clusterProviderItem, err := capiutils.GetCapiClusterProviderConfig(manifest, config_name)
	if err != nil {
		return
	}
	if clusterProviderItem.CertManager == nil {
		err = eputils.GetError("errCertMgrCfg")
		return
	}
	config.CertManager = &CertManagerConfig{Version: clusterProviderItem.CertManager.Version}

	for _, binariesItem := range clusterProviderItem.Binaries {
		if binariesItem == nil {
			err = eputils.GetError("errBinaries")
			return
		}
		if binariesItem.Name == "oras" {
			config.Binaries = &BinariesConfig{Name: binariesItem.Name, Version: binariesItem.Version}
		}
		for _, provider := range clusterProviderItem.Providers {
			if provider.Parameters == nil {
				log.Errorf("provider %s miss config", provider.Name)
				err = eputils.GetError("errProvConfig")
				return
			}

			pconfig := ProviderConfig{Name: provider.Name, Version: provider.Parameters.Version, Label: provider.Parameters.ProviderLabel}
			switch provider.ProviderType {
			case "CoreProvider":
				config.CoreProvider = &pconfig
			case "BootstrapProvider":
				config.BootstrapProvider = &pconfig
			case "ControlPlaneProvider":
				config.ControlPlaneProvider = &pconfig
			case "InfrastructureProvider":
				config.InfrastructureProvider = &pconfig
			}
		}
	}

	if config.CertManager == nil {
		err = eputils.GetError("errBinariesLost")
		return
	}
	if config.CoreProvider == nil || config.BootstrapProvider == nil || config.ControlPlaneProvider == nil || config.InfrastructureProvider == nil {
		err = eputils.GetError("errProviderLost")
		return
	}

	return
}

func generateLocalPRPath(file *pluginapi.FilesItems0, config *ClusterCtlConfig) (target string) {
	file_name := path.Base(file.URL)

	if strings.Contains(file.Mirrorurl, config.CoreProvider.Label) {
		target = filepath.Join(config.RuntimeDir, config.CoreProvider.Label, config.CoreProvider.Version, file_name)
	} else if strings.Contains(file.Mirrorurl, config.BootstrapProvider.Label) {
		target = filepath.Join(config.RuntimeDir, config.BootstrapProvider.Label, config.BootstrapProvider.Version, file_name)
	} else if strings.Contains(file.Mirrorurl, config.ControlPlaneProvider.Label) {
		target = filepath.Join(config.RuntimeDir, config.ControlPlaneProvider.Label, config.ControlPlaneProvider.Version, file_name)
	} else if strings.Contains(file.Mirrorurl, config.InfrastructureProvider.Label) {
		target = filepath.Join(config.RuntimeDir, config.InfrastructureProvider.Label, config.InfrastructureProvider.Version, file_name)
	} else if strings.Contains(file.Mirrorurl, "cert-manager") {
		target = filepath.Join(config.RuntimeDir, "cert-manager", config.CertManager.Version, file_name)
	} else if strings.Contains(file.Mirrorurl, "oras") {
		target = filepath.Join(config.RuntimeDir, "oras", config.Binaries.Name, config.Binaries.Version, file_name)
	}
	return
}

func generateLocalProviderRepo(files *pluginapi.Files, clusterctl_config *ClusterCtlConfig) error {

	for _, file := range files.Files {
		target := generateLocalPRPath(file, clusterctl_config)
		if target == "" {
			continue
		}

		err := repoutils.PullFileFromRepo(target, file.Mirrorurl)
		if err != nil {
			log.Errorf("%s, %s", err, target)
			return eputils.GetError("errPullFile")
		}
	}

	return nil
}

func generateClusterctlConfig(config *ClusterCtlConfig) (targetPath string, err error) {
	tpconfig := template.Must(template.New("pconfigyaml").Parse(template_config_yaml))

	if err = eputils.CreateFolderIfNotExist(config.RuntimeDir); err != nil {
		return
	}
	targetPath = filepath.Join(config.RuntimeDir, "config.yaml")

	var content bytes.Buffer
	if err = tpconfig.Execute(&content, *config); err != nil {
		return
	}

	if err = eputils.WriteStringToFile(content.String(), targetPath); err != nil {
		return
	}

	return
}

func launchCapiProvider(ep_params *pluginapi.EpParams, config *ClusterCtlConfig, config_path string, input_files *pluginapi.Files) error {
	bin := filepath.Join(ep_params.Runtimebin, "clusterctl")
	err := repoutils.PullFileFromRepo(bin, input_files.Files[0].Mirrorurl)
	if err != nil {
		log.Errorf("%s", err)
		return eputils.GetError("errPullFile")
	}
	err = os.Chmod(bin, 0700)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	mClusterConfig := capiutils.GetManagementClusterKubeconfig(ep_params)
	if mClusterConfig != "" {
		cmd = exec.Command(bin, "init", "--core", config.CoreProvider.Name+":"+config.CoreProvider.Version, "--bootstrap", config.BootstrapProvider.Name+":"+config.BootstrapProvider.Version, "--control-plane", config.ControlPlaneProvider.Name+":"+config.ControlPlaneProvider.Version, "--infrastructure", config.InfrastructureProvider.Name+":"+config.InfrastructureProvider.Version, "--config", config_path, "--kubeconfig", mClusterConfig)
	} else {
		cmd = exec.Command(bin, "init", "--core", config.CoreProvider.Name+":"+config.CoreProvider.Version, "--bootstrap", config.BootstrapProvider.Name+":"+config.BootstrapProvider.Version, "--control-plane", config.ControlPlaneProvider.Name+":"+config.ControlPlaneProvider.Version, "--infrastructure", config.InfrastructureProvider.Name+":"+config.InfrastructureProvider.Version, "--config", config_path)
	}

	_, err = eputils.RunCMD(cmd)
	if err != nil {
		log.Errorf("%s", err)
		return eputils.GetError("errRunClusterctlCmd")
	}

	return nil
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_cluster_manifest := input_cluster_manifest(in)
	input_files := input_files(in)

	log.Infof("Plugin: capi-provider-launch")

	if input_ep_params == nil || input_ep_params.Kitconfig == nil {
		log.Errorln("Failed to find Kitconfigs for ClusterAPI cluster.")
		return eputils.GetError("errIncorrectParam")
	}

	infra_provider, err := capiutils.GetInfraProvider(input_ep_params.Kitconfig)
	if err != nil {
		log.Errorln(err)
		return eputils.GetError("errGetCAPIInfraProv")
	}

	if input_cluster_manifest == nil || input_cluster_manifest.CapiClusterProviders == nil {
		log.Errorln("Failed to find manifest for ClusterAPI cluster.")
		return eputils.GetError("errCAPIManifest")
	}

	clusterCtlConfig, err := generateClusterCtlConfig(input_cluster_manifest.CapiClusterProviders, infra_provider, input_ep_params)
	if err != nil {
		log.Errorln(err)
		return eputils.GetError("errCAPIProvider")
	}

	if err := generateLocalProviderRepo(input_files, &clusterCtlConfig); err != nil {
		log.Errorln(err)
		return eputils.GetError("errGenProvRepoCctl")
	}

	configFilePath, err := generateClusterctlConfig(&clusterCtlConfig)
	if err != nil {
		log.Errorln(err)
		return eputils.GetError("errGenCfgClusterctl")
	}

	if err := launchManagementCluster(input_ep_params, input_cluster_manifest, input_files); err != nil {
		log.Errorln(err)
		return eputils.GetError("errLaunchMgmtClster")
	}

	if err := launchCapiProvider(input_ep_params, &clusterCtlConfig, configFilePath, input_files); err != nil {
		log.Errorf("%s", err)
		return eputils.GetError("errInitClusterctl")
	}

	return nil
}
