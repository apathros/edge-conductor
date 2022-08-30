/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package capihostprovision

import (
	cmapi "github.com/intel/edge-conductor/pkg/api/certmgr"
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	certmgr "github.com/intel/edge-conductor/pkg/certmgr"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	capiutils "github.com/intel/edge-conductor/pkg/eputils/capiutils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
	serviceutil "github.com/intel/edge-conductor/pkg/eputils/service"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	TIMEOUT     = 360
	WAIT_10_SEC = 10
	WAIT_1_SEC  = 1

	IRONICCSR               = "config/certificate/ironic/ironic-csr.json"
	IRONICINSPECTORCSR      = "config/certificate/ironic/ironic-inspector-csr.json"
	MARIADBCSR              = "config/certificate/ironic/mariadb-csr.json"
	IRONICCACSR             = "config/certificate/ironic/ironic-ca-csr.json"
	IRONICCACERTFILE        = "cert/pki/ironic/ironic-ca.pem"
	IRONICCAKEYFILE         = "cert/pki/ironic/ironic-ca-key.pem"
	IRONICCERTFILE          = "cert/pki/ironic/ironic.pem"
	IRONICKEYFILE           = "cert/pki/ironic/ironic-key.pem"
	IRONICINSPECTORCERTFILE = "cert/pki/ironicinspector/ironicinspector.pem"
	IRONICINSPECTORKEYFILE  = "cert/pki/ironicinspector/ironicinspector-key.pem"
	MARIADBCERTFILE         = "cert/pki/mariadb/mariadb.pem"
	MARIADBKEYFILE          = "cert/pki/mariadb/mariadb-key.pem"
)

func ironic_tls_setup(ep_params *pluginapi.EpParams, ironic_host_ip string) error {
	mariadb_host_ip := "127.0.0.1" //Need to be checked

	initcertsCa := &cmapi.CertificateCa{
		Csr:  IRONICCACSR,
		Cert: IRONICCACERTFILE,
		Key:  IRONICCAKEYFILE,
	}

	ironiccerts := cmapi.Certificate{
		Name: "ironic",
		Ca:   initcertsCa,
		Server: &cmapi.CertificateServer{
			Csr:  IRONICCSR,
			Cert: IRONICCERTFILE,
			Key:  IRONICKEYFILE,
		},
	}
	ironicinspectorcerts := cmapi.Certificate{
		Name: "ironicinspector",
		Ca:   initcertsCa,
		Server: &cmapi.CertificateServer{
			Csr:  IRONICINSPECTORCSR,
			Cert: IRONICINSPECTORCERTFILE,
			Key:  IRONICINSPECTORKEYFILE,
		},
	}
	mariadbcerts := cmapi.Certificate{
		Name: "mariadb",
		Ca:   initcertsCa,
		Server: &cmapi.CertificateServer{
			Csr:  MARIADBCSR,
			Cert: MARIADBCERTFILE,
			Key:  MARIADBKEYFILE,
		},
	}
	// check and gen certs
	if err := certmgr.GenCertAndConfig(ironiccerts, ironic_host_ip); err != nil {
		log.Error(err)
		return err
	}
	if err := certmgr.GenCertAndConfig(ironicinspectorcerts, ironic_host_ip); err != nil {
		log.Error(err)
		return err
	}
	if err := certmgr.GenCertAndConfig(mariadbcerts, mariadb_host_ip); err != nil {
		log.Error(err)
		return err
	}

	// Ensure that the MariaDB key file allow a non-owned user to read.
	if eputils.FileExists(MARIADBKEYFILE) {
		err := os.Chmod(MARIADBKEYFILE, 0604)
		if err != nil {
			return err
		}
	}

	return nil
}

func launchIronicContainers(ep_params *pluginapi.EpParams, workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	var err error

	dstFile := filepath.Join(workFolder, "ironic-containers.yaml")
	err = capiutils.TmplFileRendering(tmpl, workFolder, clusterConfig.BaremetelOperator.IronicContainers, dstFile)
	if err != nil {
		log.Errorf("Failed to render %s, %v", clusterConfig.BaremetelOperator.IronicContainers, err)
		return err
	}

	var ironicContainers pluginapi.Containers
	err = eputils.LoadSchemaStructFromYamlFile(&ironicContainers, dstFile)
	if err != nil {
		log.Errorf("Load capi cluster config failed, %v", err)
		return err
	}

	for _, c := range ironicContainers.Containers {
		if c.Name == "ipa-downloader" {
			continue
		}

		if err = docker.DockerRun(c); err != nil {
			log.Errorf("Container %s run fail, %v", c.Name, err)
			return err
		}
	}

	return nil
}

//nolint: dupl   //TODO: Need fix this error in v0.6
func launchBmo(ep_params *pluginapi.EpParams, workFolder, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	var err error

	dstFile := filepath.Join(workFolder, "bmo-manifest.yaml")
	err = capiutils.TmplFileRendering(tmpl, workFolder, clusterConfig.BaremetelOperator.URL, dstFile)
	if err != nil {
		log.Errorf("Failed to render %s, %v", clusterConfig.BaremetelOperator.URL, err)
		return err
	}

	name := "baremetal-operator"
	deployer := serviceutil.NewYamlDeployer(name, clusterConfig.WorkloadCluster.Namespace, dstFile)
	err = deployer.YamlInstall(management_kubeconfig)
	defer func() {
		err := os.RemoveAll(dstFile)
		if err != nil {
			log.Errorf("Fail to remove file, %v", err)
		}
	}()

	if err != nil {
		log.Errorf("Baremetal operator deploy fail, %v", err)
		return err
	}

	return nil
}

//nolint: dupl   //TODO: Need fix this error in v0.6
func makeBmHosts(ep_params *pluginapi.EpParams, workFolder, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	var err error

	dstFile := filepath.Join(workFolder, "bmhost.yaml")
	err = capiutils.TmplFileRendering(tmpl, workFolder, clusterConfig.BaremetelOperator.Bmhost, dstFile)
	if err != nil {
		log.Errorf("Failed to render %s, %v", clusterConfig.BaremetelOperator.Bmhost, err)
		return err
	}

	name := "baremetal-host"
	deployer := serviceutil.NewYamlDeployer(name, clusterConfig.WorkloadCluster.Namespace, dstFile)
	err = deployer.YamlInstall(management_kubeconfig)
	defer func() {
		err := os.RemoveAll(dstFile)
		if err != nil {
			log.Errorf("Fail to remove file, %v", err)
		}
	}()

	if err != nil {
		log.Errorf("Bm-host deploy fail, %v", err)
		return err
	}

	return nil
}

func checkBmHosts(ep_params *pluginapi.EpParams, workFolder, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	ready := false
	count := 0

	for count < TIMEOUT {
		cmd := exec.Command(ep_params.Workspace+"/kubectl", "get", "bmh", "-n", clusterConfig.WorkloadCluster.Namespace, "--kubeconfig", management_kubeconfig)

		outputStr, err := eputils.RunCMD(cmd)
		if err != nil {
			log.Errorf("Failed to get baremetal hosts. %v", err)
			return err
		}

		reAvaliable := regexp.MustCompile(`.*\savailable\s.*`)
		matchesOutput := reAvaliable.FindAllStringSubmatch(outputStr, -1)

		if len(matchesOutput) < 1 {
			log.Infof("sleep %d sec", WAIT_10_SEC)
			time.Sleep(WAIT_10_SEC * time.Second)
			count++
		} else {
			ready = true
			break
		}
	}

	if !ready {
		log.Errorf("Node is not ready, please check")
		return eputils.GetError("errNodeNotReady")
	}

	return nil
}

func metal3HostProvision(ep_params *pluginapi.EpParams, workFolder, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	var err error

	err = ironic_tls_setup(ep_params, tmpl.CapiSetting.IronicConfig.IronicProvisionIP)
	if err != nil {
		log.Errorf("Ironic tls setup failed, %v", err)
		return err
	}

	err = launchBmo(ep_params, workFolder, management_kubeconfig, clusterConfig, tmpl)
	if err != nil {
		log.Errorf("Launch baremetal operator fail")
		return err
	}

	err = DeploymentReady(management_kubeconfig, "baremetal-operator-system", "baremetal-operator-controller-manager")
	if err != nil {
		log.Errorf("Bmo deployment launch fail, %v", err)
		return err
	}

	err = launchIronicContainers(ep_params, workFolder, clusterConfig, tmpl)
	if err != nil {
		log.Errorf("Launch ironic containers fail")
		return err
	}

	err = DeploymentReady(management_kubeconfig, "baremetal-operator-system", "baremetal-operator-controller-manager")
	if err != nil {
		log.Errorf("Bmo deployment launch fail, %v", err)
		return err
	}

	err = makeBmHosts(ep_params, workFolder, management_kubeconfig, clusterConfig, tmpl)
	if err != nil {
		log.Errorf("Make bmHost fail")
		return err
	}

	err = checkBmHosts(ep_params, workFolder, management_kubeconfig, clusterConfig, tmpl)
	if err != nil {
		log.Errorf("Failed to get available baremetal-host, %v", err)
		return err
	}

	return nil
}
