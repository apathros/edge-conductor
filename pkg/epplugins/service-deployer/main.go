/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package servicedeployer

import (
	"fmt"
	epplugins "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	kubeutils "github.com/intel/edge-conductor/pkg/eputils/kubeutils"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	serviceutil "github.com/intel/edge-conductor/pkg/eputils/service"
	"github.com/intel/edge-conductor/pkg/executor"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	epConfigMapName          = "edgeconductor-service"
	epNamespace              = "edgeconductor"
	epFieldManagerName       = "Edge Conductor"
	epConfigmapResourcesName = "serviceOverrideHash"
)

func findService(serviceName string, serviceConfig *epplugins.Serviceconfig) *epplugins.Component {
	if serviceConfig == nil {
		return nil
	}
	for _, service := range serviceConfig.Components {
		if service.Name == serviceName {
			return service
		}
	}
	return nil
}

func getExpectedRevision(configmap kubeutils.ConfigMapWrapper, servicename string) string {
	expectRevision := ""
	appliedService := &epplugins.Component{}
	data := configmap.GetData()
	if err := eputils.LoadSchemaStructFromYaml(appliedService,
		data[servicename]); err != nil {
		log.Errorln("Failed to load service ConfigMap:", err)
	} else {
		expectRevision = appliedService.Revision
	}
	return expectRevision
}

func getExpectedChartHash(configmap kubeutils.ConfigMapWrapper, servicename string) string {
	expectHash := ""
	appliedService := &epplugins.Component{}
	data := configmap.GetData()
	if err := eputils.LoadSchemaStructFromYaml(appliedService,
		data[servicename]); err != nil {
		log.Errorln("Failed to load service ConfigMap:", err)
	} else {
		expectHash = appliedService.Hash
	}

	return expectHash
}

func getExpectedOverrideHash(configmap kubeutils.ConfigMapWrapper, servicename string) string {
	expectServiceOverrideHash := ""
	appliedService := &epplugins.Component{}
	data := configmap.GetData()
	if err := eputils.LoadSchemaStructFromYaml(appliedService,
		data[servicename]); err != nil {

		log.Errorln("Failed to load service ConfigMap:", err)
	} else {
		for _, resource := range appliedService.Resources {
			if resource.Name == epConfigmapResourcesName {
				expectServiceOverrideHash = resource.Value
				break
			}
		}
	}
	return expectServiceOverrideHash
}

func getRevision(deployer serviceutil.HelmDeployerWrapper, runtime_kubeconfig string, serviceName string) (string, error) {
	status, rev := deployer.HelmStatus(runtime_kubeconfig)
	if rev == 0 {
		log.Errorf("Helm service %s with wrong status: %s", serviceName, status)
		return "", eputils.GetError("errWrongStatus")
	}
	revision := fmt.Sprintf("%d", rev)
	return revision, nil
}

func updateConfigmap(service *epplugins.Component, configMap kubeutils.ConfigMapWrapper, localValueSha256Str string) error {
	item := epplugins.ComponentResourcesItems0{Name: epConfigmapResourcesName, Value: localValueSha256Str}
	// Add or update the service in ConfigMap
	service.Resources = append(service.Resources, &(item))
	data, err := eputils.SchemaStructToYaml(service)
	if err != nil {
		log.Errorln(err)
		return err
	}
	if err := configMap.RenewData(service.Name, data); err != nil {
		log.Errorln(err)
		return err
	}
	return nil
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_serviceconfig := input_serviceconfig(in)

	runtime_kubeconfig := input_ep_params.Kubeconfig

	tmpDir := filepath.Join(input_ep_params.Runtimedir, "tmp")
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.Errorln("failed to remove", tmpDir, err)
		}
	}()

	// Create Namespace
	err := kubeutils.CreateNamespace(runtime_kubeconfig, epNamespace)
	if err != nil {
		return err
	}
	// Get the ConfigMap of applied service list.
	serviceConfigMap, err := kubeutils.NewConfigMap(
		epNamespace, epConfigMapName, epFieldManagerName, runtime_kubeconfig)
	if err != nil {
		return err
	}
	err = serviceConfigMap.Get()
	if err != nil {
		log.Infoln("ConfigMap", epConfigMapName, "not found on cluster, will create a new one.")
		if err := serviceConfigMap.New(); err != nil {
			log.Errorln("Failed to get or create ConfigMap", epConfigMapName)
			return err
		}
	}

	// If an applied service is not in current service list, uninstall it.
	for _, yml := range serviceConfigMap.GetData() {
		appliedService := &epplugins.Component{}
		err := eputils.LoadSchemaStructFromYaml(appliedService, yml)
		if err != nil {
			log.Errorln("Failed to load service ConfigMap:", err)
			return err
		}

		if appliedService.Type == "yaml" {
			s := findService(appliedService.Name, input_serviceconfig)
			if s == nil {
				log.Infoln(appliedService.Name, "is not in current service list, will be installed.")
				namespace := appliedService.Namespace
				if len(namespace) <= 0 {
					namespace = "default"
				}
				targetFile := filepath.Join(tmpDir, appliedService.Name+".yml")
				if err := repoutils.PullFileFromRepo(targetFile, appliedService.URL); err != nil {
					log.Errorln("Failed to pull file", appliedService.URL)
					return err
				}
				err := eputils.FileTemplateConvert(targetFile, targetFile)
				if err != nil {
					log.Errorln("File Template Convert Failed:", err)
				}
				// The service is not in current list
				deployer := serviceutil.NewYamlDeployer(appliedService.Name, namespace, targetFile)
				// Uninstall the service
				if err := deployer.YamlUninstall(runtime_kubeconfig); err != nil {
					if strings.Contains(fmt.Sprintln(err), "NotFound") {
						log.Warnln("Resource not found when uninstalling", deployer.GetName())
					} else {
						log.Errorln(err)
						return err
					}
				}
				// Remove the data entry from ConfigMap
				if err := serviceConfigMap.RemoveData(appliedService.Name); err != nil {
					log.Errorln(err)
					return err
				}
				log.Infoln(deployer.GetName(), "uninstalled/removed.")
			}
		} else if appliedService.Type == "helm" {
			if s := findService(appliedService.Name, input_serviceconfig); s == nil {
				namespace := appliedService.Namespace
				if len(namespace) <= 0 {
					namespace = "default"
				}

				var localChart string
				if appliedService.URL != "" {
					localChart = filepath.Join(tmpDir, appliedService.Name+".tgz")
					if err := repoutils.PullFileFromRepo(localChart, appliedService.URL); err != nil {
						log.Errorln("Failed to pull file", appliedService.URL, "to", localChart)
						return err
					}
				} else {
					localChart = ""
				}
				var localValue string
				if appliedService.Chartoverride != "" {
					localValue = filepath.Join(tmpDir, appliedService.Name+".yml")
					if err := repoutils.PullFileFromRepo(localValue, appliedService.Chartoverride); err != nil {
						log.Errorln("Failed to pull file", appliedService.Chartoverride, "to", localValue)
						return err
					}
					errFileTemplateConvert := eputils.FileTemplateConvert(localValue, localValue)
					if errFileTemplateConvert != nil {
						log.Errorln("File Template Convert Failed:", errFileTemplateConvert)
					}
				} else {
					localValue = ""
				}

				// The service is not in current list
				deployer := serviceutil.NewHelmDeployer(
					appliedService.Name,
					namespace,
					localChart,
					localValue,
				)
				if status, rev := deployer.HelmStatus(runtime_kubeconfig); status == serviceutil.HELM_STATUS_UNKNOWN {
					// Unknown Status
					log.Errorln(deployer.GetName(), "current status unknown, need to check cluster status.")
					log.Errorf("Helm service %s status unknown", deployer.GetName())
					return eputils.GetError("errUnknownStatus")
				} else if status == serviceutil.HELM_STATUS_NOT_DEPLOYED {
					// Helm is not deployed, nothing to do.
				} else if status == serviceutil.HELM_STATUS_DEPLOYED {
					// Helm is already deployed, need to check if the revision is as expected.
					expectRevision := getExpectedRevision(serviceConfigMap, appliedService.Name)
					if expectRevision == fmt.Sprintf("%d", rev) {
						log.Infof("Release %s rev.%d is deployed, will uninstall the service.", appliedService.Name, rev)
						// Uninstall the service
						if err := deployer.HelmUninstall(runtime_kubeconfig); err != nil {
							log.Errorln(err)
							return err
						}
					} else {
						log.Warnf("Expect %s rev.%s but rev.%d found.", deployer.GetName(), expectRevision, rev)
						continue
						// TODO: Need to decide whether to return an error here.
						// return errors.New(fmt.Sprintf("Expect %s rev.%s but rev.%d found.", deployer.GetName(), expectRevision, rev))
					}
				} else {
					// Wrong Status Found
					// Uninstall the service
					log.Infof("Release %s is in a wrong status %s, will uninstall the service.", deployer.GetName(), status)
					if err := deployer.HelmUninstall(runtime_kubeconfig); err != nil {
						log.Errorf("Failed to uninstall %s, which is previously in a wrong status %s, please uninstall it manually.", deployer.GetName(), status)
						log.Errorln(err)
					}
				}

				// Remove the data entry from ConfigMap
				if err := serviceConfigMap.RemoveData(appliedService.Name); err != nil {
					log.Errorln(err)
					return err
				}
			}
		}
	}

	// Install/upgrade all services in current list.
	for _, service := range input_serviceconfig.Components {
		if service.Type == "yaml" {
			if service.Executor != nil && service.Executor.Deploy != "" {
				log.Errorf("No DCE deploy spec supported for %s %s", service.Type, service.Name)
				return eputils.GetError("errWrongOperation")
			}

			log.Infof("Yaml service %s will be deployed.", service.Name)

			// Create namespace if specified.
			namespace := service.Namespace
			if len(namespace) <= 0 {
				namespace = "default"
			}
			if namespace != "default" {
				err := kubeutils.CreateNamespace(runtime_kubeconfig, namespace)
				if err != nil {
					return err
				}
			}
			targetFile := filepath.Join(tmpDir, service.Name+".yml")
			err := repoutils.PullFileFromRepo(targetFile, service.URL)
			if err != nil {
				log.Errorln("Failed to pull file", service.URL)
				return err
			}
			// Create deployer
			errFileTemplateConvert := eputils.FileTemplateConvert(targetFile, targetFile)
			if errFileTemplateConvert != nil {
				log.Errorln("File Template Convert Failed:", errFileTemplateConvert)
			}
			wait := &serviceutil.YamlWait{Timeout: 0}
			if service.Wait != nil && service.Wait.Timeout != 0 {
				wait.Timeout = service.Wait.Timeout
				log.Infof("service (%s) will wait", service.Name)
			}
			deployer := serviceutil.NewYamlDeployer(service.Name, namespace, targetFile, wait)

			// Install the service
			err = deployer.YamlInstall(runtime_kubeconfig)
			if err != nil {
				log.Errorln(err)
				return err
			}
			log.Infoln(deployer.GetName(), "successfully installed.")
			// Add or update the service in ConfigMap
			data, err := eputils.SchemaStructToYaml(service)
			if err != nil {
				log.Errorln(err)
				return err
			}
			err = serviceConfigMap.RenewData(service.Name, data)
			if err != nil {
				log.Errorln(err)
				return err
			}
		} else if service.Type == "helm" {
			if service.Executor != nil && service.Executor.Deploy != "" {
				log.Errorf("No DCE deploy spec supported for %s %s", service.Type, service.Name)
				return eputils.GetError("errWrongOperation")
			}

			log.Infof("Helm service %s will be deployed.", service.Name)
			namespace := service.Namespace
			if len(namespace) <= 0 {
				namespace = "default"
			}
			if namespace != "default" {
				err := kubeutils.CreateNamespace(runtime_kubeconfig, namespace)
				if err != nil {
					return err
				}
			}
			// Prepare tls secrets
			err := serviceutil.GenSvcSecretFromTLSExtension(input_ep_params.Extensions, service.Name, namespace, runtime_kubeconfig)
			if err != nil {
				return err
			}

			var localChart string
			var localChartSha256Str string
			if service.URL != "" {
				localChart = filepath.Join(tmpDir, service.Name+".tgz")
				if err := repoutils.PullFileFromRepo(localChart, service.URL); err != nil {
					log.Errorln("Failed to pull file", service.URL, "to", localChart)
					return err
				}
				if len(service.Hash) == 0 {
					if localChartSha256Str, err = eputils.GenFileSHA256(localChart); err != nil {
						log.Errorln("Failed to generate SHA256 hash code for helm charts of", service.Name)
						return err
					} else {
						service.Hash = localChartSha256Str
					}
				}
			} else {
				localChart = ""
				localChartSha256Str = ""
			}
			var localValue string
			var localValueSha256Str string
			if service.Chartoverride != "" {
				localValue = filepath.Join(tmpDir, service.Name+".yml")
				if err := repoutils.PullFileFromRepo(localValue, service.Chartoverride); err != nil {
					log.Errorln("Failed to pull file", service.Chartoverride, "to", localValue)
					return err
				}
				err := eputils.FileTemplateConvert(localValue, localValue)
				if err != nil {
					log.Errorln("File Template Convert Failed:", err)
				}
				if localValueSha256Str, err = eputils.GenFileSHA256(localValue); err != nil {
					return err
				}
			} else {
				localValue = ""
				localValueSha256Str = ""
			}
			deployer := serviceutil.NewHelmDeployer(
				service.Name,
				namespace,
				localChart,
				localValue,
			)

			if status, rev := deployer.HelmStatus(runtime_kubeconfig); status == serviceutil.HELM_STATUS_UNKNOWN {
				// Unknown Status

				log.Warnln(deployer.GetName(), "current status unknown, need to check cluster status.")
				log.Warningf("Helm service %s status unknown", deployer.GetName())
				return eputils.GetError("errUnknownStatus")
			} else if status == serviceutil.HELM_STATUS_NOT_DEPLOYED {
				// Helm is not deployed, need a new install.
				wait := false
				timeout := 0
				if service.Wait != nil && service.Wait.Timeout != 0 {
					wait = true
					timeout = int(service.Wait.Timeout)
					log.Infof("service (%s) will wait", service.Name)
				}
				if err := deployer.HelmInstall(runtime_kubeconfig, serviceutil.WithWaitAndTimeout(wait, timeout)); err != nil {
					// Known issue for wait crd, WA to deloy 2nd time
					if status, _ := deployer.HelmStatus(runtime_kubeconfig); status == serviceutil.HELM_STATUS_NOT_DEPLOYED {
						if err = deployer.HelmInstall(runtime_kubeconfig, serviceutil.WithWaitAndTimeout(wait, timeout)); err != nil {
							log.Errorln(" 2nd Deploy Error met: ", err)
							return err
						}
					} else {
						log.Errorln(err)
						return err
					}
				}
			} else if status == serviceutil.HELM_STATUS_DEPLOYED {
				// Helm is already deployed, need to check if the revision is as expected.
				expectRevision := getExpectedRevision(serviceConfigMap, service.Name)
				expectChartHash := getExpectedChartHash(serviceConfigMap, service.Name)
				expectOverrideHash := getExpectedOverrideHash(serviceConfigMap, service.Name)
				if expectRevision == fmt.Sprintf("%d", rev) {
					if expectChartHash == service.Hash && expectOverrideHash == localValueSha256Str {
						log.Infof("The current %s hash has not changed, no need to upgrade", service.Name)
						continue
					}
					log.Infof("Release %s rev.%d is already deployed, will upgrade the service.", service.Name, rev)
					if err := deployer.HelmUpgrade(runtime_kubeconfig); err != nil {
						log.Errorln(err)
						return err
					}
				} else {
					log.Warnf("Expect %s rev.%s but rev.%d found.", deployer.GetName(), expectRevision, rev)
					continue
					// TODO: Need to decide whether to return an error here.
					// return errors.New(fmt.Sprintf("Expect %s rev.%s but rev.%d found.", deployer.GetName(), expectRevision, rev))
				}
			} else {
				// Wrong Status Found
				// As there's a wrong status found, report error.
				log.Errorf("%s is in a wrong status %s, please remove it from the selector list and re-run the \"service build/deploy\"", deployer.GetName(), status)
				return eputils.GetError("errServiceStatus")
			}
			if service.Revision, err = getRevision(deployer, runtime_kubeconfig, service.Name); err != nil {
				return err
			}
			//add Rescources
			if err := updateConfigmap(service, serviceConfigMap, localValueSha256Str); err != nil {
				return err
			}
		} else if service.Type == "dce" {
			if service.Executor.Deploy != "" {
				log.Infof("DCE service %s will be deployed.", service.Name)
				err := executor.Run(service.Executor.Deploy, input_ep_params, service)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
