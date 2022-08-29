/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package service

import (
	"bytes"
	eputils "ep/pkg/eputils"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
)

type K8SResource struct {
	Kind     string `yaml:"kind"`
	MetaData struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Spec struct {
		Template struct {
			MetaData struct {
				Labels map[string]string `yaml:"labels"`
			} `yaml:"metadata"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

func (h *YamlDeployer) WaitResource(k8sres *K8SResource, loc_kubeconfig string) error {
	var res string
	var cond string
	var labels string
	var name string
	var cmd *exec.Cmd
	if h.Wait.Timeout == 0 {
		log.Infoln("Timeout is 0")
		return nil
	}
	timeout := strconv.FormatInt(h.Wait.Timeout, 10)
	log.Infof("Resource %s Wait", k8sres.Kind)
	switch k8sres.Kind {
	case "Job":
		res = "Job"
		name = k8sres.MetaData.Name
		cond = "--for=condition=Complete"
		cmd = exec.Command(
			"./kubectl",
			"--kubeconfig", loc_kubeconfig,
			"wait",
			cond,
			res,
			name,
			fmt.Sprintf("--timeout=%ss", timeout),
			"-n", h.Namespace)

	case "Deployment": //change to check Pod of Deployment
		res = "Pod"
		cond = "--for=condition=Ready"
		llen := len(k8sres.Spec.Template.MetaData.Labels)
		i := 0
		for k, v := range k8sres.Spec.Template.MetaData.Labels {
			labels += fmt.Sprintf("%s=%s", k, v)
			if i != llen-1 {
				labels += ","
			}
			i += 1
		}
		if labels != "" {
			labels = "-l " + labels
		}
		cmd = exec.Command(
			"./kubectl",
			"--kubeconfig", loc_kubeconfig,
			"wait",
			cond,
			res,
			labels,
			fmt.Sprintf("--timeout=%ss", timeout),
			"-n", h.Namespace)

	default:
		log.Infof("Skip Resource %s Wait", k8sres.Kind)
		return nil
	}

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("KUBECONFIG=%s", loc_kubeconfig),
		fmt.Sprintf("YAML=%s", h.LocYaml),
		fmt.Sprintf("NAMESPACE=%s", h.Namespace),
	)
	log.Debug("Kube wait cmd is: ", cmd.String())
	_, err := eputils.RunCMD(cmd)
	if err != nil {
		return err
	}
	return nil
}

type YamlWait struct {
	Timeout int64
}

type YamlDeployer struct {
	LocYaml   string
	Name      string
	Namespace string
	Wait      *YamlWait
}

func NewYamlDeployer(name, namespace, yamlfile string, para ...interface{}) YamlDeployerWrapper {

	// only take the first wait condition
	if len(para) != 0 {
		wait, ok := para[0].(*YamlWait)
		if ok {
			return &YamlDeployer{
				LocYaml:   yamlfile,
				Name:      name,
				Namespace: namespace,
				Wait:      wait,
			}
		}
		log.Warning("Parameter is wrong, skip wait setting")
	}
	return &YamlDeployer{
		LocYaml:   yamlfile,
		Name:      name,
		Namespace: namespace,
		Wait:      nil,
	}
}

func (h *YamlDeployer) __run_kube_script(loc_kubeconfig, op string) error {
	cmd := exec.Command(
		"./kubectl",
		"--kubeconfig", loc_kubeconfig,
		op,
		"-f", h.LocYaml)

	cmd.Env = append(os.Environ(),
		fmt.Sprintf("KUBECONFIG=%s", loc_kubeconfig),
		fmt.Sprintf("YAML=%s", h.LocYaml),
		fmt.Sprintf("NAMESPACE=%s", h.Namespace),
	)

	_, err := eputils.RunCMD(cmd)

	return err
}

func (h *YamlDeployer) GetName() string {
	return h.Name
}

func (h *YamlDeployer) YamlInstall(loc_kubeconfig string) error {
	log.Infoln("Kube Apply YAML:", h.Name)

	err := h.__run_kube_script(loc_kubeconfig, "apply")
	if err != nil {
		return err
	}
	if h.Wait != nil && h.Wait.Timeout != 0 {
		res := K8SResource{}
		data, err := ioutil.ReadFile(h.LocYaml)
		if err != nil {
			return err
		}
		r := bytes.NewReader(data)
		dec := yaml.NewDecoder(r)
		for dec.Decode(&res) == nil {
			if err := h.WaitResource(&res, loc_kubeconfig); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *YamlDeployer) YamlUninstall(loc_kubeconfig string) error {
	log.Infoln("Kube Delete YAML:", h.Name)

	return h.__run_kube_script(loc_kubeconfig, "delete")
}
