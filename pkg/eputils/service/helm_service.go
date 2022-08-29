/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package service

import (
	"bytes"
	"fmt"
	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"time"
)

const (
	HELM_STATUS_PENDING      = "Pending"
	HELM_STATUS_DEPLOYED     = "Deployed"
	HELM_STATUS_NOT_DEPLOYED = "Not Deployed"
	HELM_STATUS_UNKNOWN      = "Unknown"
)

var gActionConfig *action.Configuration

// HelmDeployer: Class for Helm deployment.
//   LocCharts:  Location of the charts. Can be a local file or a remote URL.
//   LocValues:  Location of the value file. Can be a local file or a remote URL.
//   Name:       Name of the release.
//   Namespace:  Namespace for the release to run.
//
type HelmDeployer struct {
	LocCharts string
	LocValues string
	Name      string
	Namespace string
}

func NewHelmDeployer(name, namespace, charts, values string) HelmDeployerWrapper {
	return &HelmDeployer{
		LocCharts: charts,
		LocValues: values,
		Name:      name,
		Namespace: namespace,
	}
}

// initHelm to initialize helm actionConfig.
func initHelm(kubeconfig, namespace string) error {
	gActionConfig = nil

	// Get kubeconfig
	kubeconfig_abs, err := filepath.Abs(kubeconfig)
	if err != nil {
		log.Errorln("ERROR: Failed to find kubeconfig.", err)
		return err
	}
	// Init action configuration
	gActionConfig = new(action.Configuration)
	if err := gActionConfig.Init(
		kube.GetConfig(kubeconfig_abs, "", namespace),
		namespace,
		os.Getenv("HELM_DRIVER"),
		func(format string, v ...interface{}) {
			fmt.Printf(format, v)
		}); err != nil {
		return err
	}
	return nil
}

// readFile to load a remote file with a url, or a file from the local directory.
func readFile(filePath string) ([]byte, error) {
	// Parse URL and check the scheme
	ufile, _ := url.Parse(filePath)

	// Handle http UrLs
	if ufile.Scheme == "http" || ufile.Scheme == "https" {
		buf := bytes.NewBuffer(nil)
		// Get the data
		// #nosec G107
		resp, err := http.Get(filePath)
		if err != nil {
			log.Errorln("Failed to get", filePath, err)
			return buf.Bytes(), err
		}
		defer resp.Body.Close()
		_, err = io.Copy(buf, resp.Body)
		return buf.Bytes(), err

	} else {
		// Handle local files
		return ioutil.ReadFile(filePath)
	}
}

// readValues to read values map from a url.
func readValues(fileurl string) (map[string]interface{}, error) {
	valuemap := map[string]interface{}{}

	// Read from override value file
	valuebytes, err := readFile(fileurl)
	if err != nil {
		return nil, err
	}
	// Unmarshal the values
	if err = yaml.Unmarshal(valuebytes, &valuemap); err != nil {
		return nil, err
	}

	return valuemap, nil
}

// loadChart to load a helm chart from local directory or a remote URL.
func loadChartRemote(charturl string) (*chart.Chart, error) {
	// #nosec G107
	resp, err := http.Get(charturl)
	if err != nil {
		log.Errorln("Failed to get package from url", charturl, err)
		return nil, err
	}
	defer resp.Body.Close()
	c, err := loader.LoadArchive(resp.Body)
	if err != nil {
		log.Errorln("Failed to load archive", charturl, err)
		return nil, err
	}
	return c, nil
}
func loadChartLocal(chartpos string) (*chart.Chart, error) {
	c, err := loader.Load(chartpos)
	if err != nil {
		log.Errorln("Failed to load", chartpos, err)
		return nil, err
	}
	return c, nil
}

func (h *HelmDeployer) GetName() string {
	return h.Name
}

// HelmStatus: Get status of a helm release.
//
// Parameters:
//   loc_kubeconfig:  Location of the kubeconfig file.
// Return:
//   Current status: string
//   Revision:       int
func (h *HelmDeployer) HelmStatus(loc_kubeconfig string) (string, int) {
	// Init Helm Configurations
	if err := initHelm(loc_kubeconfig, h.Namespace); err != nil {
		log.Errorln("Failed to init Helm Configuration:", err)
		return HELM_STATUS_UNKNOWN, 0
	}

	// New Client
	helmcli := action.NewStatus(gActionConfig)
	res, err := helmcli.Run(h.Name)
	if err != nil {
		if err.Error() == "release: not found" {
			return HELM_STATUS_NOT_DEPLOYED, 0
		}
		return HELM_STATUS_UNKNOWN, 0
	} else {
		if res.Info.Status.IsPending() {
			return HELM_STATUS_PENDING, res.Version
		} else if res.Info.Status == release.StatusDeployed {
			return HELM_STATUS_DEPLOYED, res.Version
		} else {
			return strings.Title(res.Info.Status.String()), res.Version
		}
	}
}

type installConfig struct {
	wait    bool
	timeout int
}

type InstallOpt func(*installConfig)

// WithWait provides the wait setting for helm install config
func WithWaitAndTimeout(wait bool, timeout int) InstallOpt {
	return func(opt *installConfig) {
		opt.wait = wait
		opt.timeout = timeout
	}
}

// HelmInstall: Install the helm charts described by the HelmDeployer
//
// Parameters:
//   loc_kubeconfig:  Location of the kubeconfig file.
//
func (h *HelmDeployer) HelmInstall(loc_kubeconfig string, opts ...InstallOpt) error {
	log.Infoln("Helm Install:", h.Name)
	log.Infoln("       Chart:", h.LocCharts)

	var conf installConfig
	for _, opt := range opts {
		opt(&conf)
	}
	// Get values
	var values map[string]interface{}
	values = nil
	if len(h.LocValues) > 0 {
		log.Infoln("       Value:", h.LocValues)

		v, err := readValues(h.LocValues)
		if err != nil {
			log.Errorln("Failed to load override file:", err)
			return err
		}
		values = v
	}

	// Init Helm Configurations
	if err := initHelm(loc_kubeconfig, h.Namespace); err != nil {
		log.Errorln("Failed to init Helm Configuration:", err)
		return err
	}

	// New Install Client
	helmcli := action.NewInstall(gActionConfig)

	// Load Chart
	chartpos, err := helmcli.ChartPathOptions.LocateChart(
		h.LocCharts, &cli.EnvSettings{})
	var ch *chart.Chart
	if err != nil {
		ch, err = loadChartRemote(h.LocCharts)
		if err != nil {
			return err
		}
	} else {
		ch, err = loadChartLocal(chartpos)
		if err != nil {
			return err
		}
	}

	dependencies := ch.Metadata.Dependencies
	if dependencies != nil {
		err = action.CheckDependencies(ch, dependencies)
		if err != nil {
			dlMgr := &downloader.Manager{
				ChartPath:  chartpos,
				SkipUpdate: false,
			}
			err = dlMgr.Update()
			if err != nil {
				log.Errorf("Failed to download dependencies charts, %v", err)
				return err
			}
			if ch, err = loader.Load(chartpos); err != nil {
				return err
			}
		}
	}

	helmcli.Namespace = h.Namespace
	helmcli.ReleaseName = h.Name
	helmcli.Timeout = time.Duration(conf.timeout) * time.Second
	helmcli.Wait = conf.wait

	rel, err := helmcli.Run(ch, values)
	if err != nil {
		log.Errorln("Failed to run Helm install:", err)
		return err
	}
	log.Infoln("")
	log.Infoln("Successfully installed release: ", rel.Name)
	return nil
}

// HelmUpgrade: Upgrade the helm charts described by the HelmDeployer
//
// Parameters:
//   loc_kubeconfig:  Location of the kubeconfig file.
//
func (h *HelmDeployer) HelmUpgrade(loc_kubeconfig string) error {
	log.Infoln("Helm Upgrade:", h.Name)
	log.Infoln("       Chart:", h.LocCharts)

	// Get values
	var values map[string]interface{}
	values = nil
	if len(h.LocValues) > 0 {
		log.Infoln("       Value:", h.LocValues)

		v, err := readValues(h.LocValues)
		if err != nil {
			log.Errorln("Failed to load override file:", err)
			return err
		}
		values = v
	}

	// Init Helm Configurations
	if err := initHelm(loc_kubeconfig, h.Namespace); err != nil {
		log.Errorln("Failed to init Helm Configuration:", err)
		return err
	}

	// New Upgrade Client
	helmcli := action.NewUpgrade(gActionConfig)

	// Load Chart
	chartpos, err := helmcli.ChartPathOptions.LocateChart(
		h.LocCharts, &cli.EnvSettings{})
	var ch *chart.Chart
	if err != nil {
		ch, err = loadChartRemote(h.LocCharts)
		if err != nil {
			return err
		}
	} else {
		ch, err = loadChartLocal(chartpos)
		if err != nil {
			return err
		}
	}

	helmcli.Namespace = h.Namespace
	rel, err := helmcli.Run(h.Name, ch, values)
	if err != nil {
		log.Errorln("Failed to run Helm upgrade:", err)
		return err
	}
	log.Infoln("")
	log.Infoln("Successfully upgraded release: ", rel.Name)
	return nil
}

// HelmUninstall: Uninstall the helm charts described by the HelmDeployer
//
// Parameters:
//   loc_kubeconfig:  Location of the kubeconfig file.
//
func (h *HelmDeployer) HelmUninstall(loc_kubeconfig string) error {
	log.Infoln("Helm Uninstall:", h.Name)

	// Init Helm Configurations
	if err := initHelm(loc_kubeconfig, h.Namespace); err != nil {
		log.Errorln("Failed to init Helm Configuration:", err)
		return err
	}

	// New Client
	helmcli := action.NewUninstall(gActionConfig)

	res, err := helmcli.Run(h.Name)
	if err != nil {
		log.Errorln("Failed to uninstall", h.Name, err)
		return err
	}
	if res != nil && res.Info != "" {
		log.Infoln(res.Info)
	}
	log.Infoln("")
	log.Infoln("Successfully uninstalled release: ", h.Name)
	return nil
}
