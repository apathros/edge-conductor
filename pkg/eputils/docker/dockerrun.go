/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//go:generate mockgen -destination=./mock/docker_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt github.com/intel/edge-conductor/pkg/eputils/docker DockerInterface

package docker

import (
	"fmt"
	"os/user"

	"github.com/docker/docker/api/types/mount"
	log "github.com/sirupsen/logrus"

	api "github.com/intel/edge-conductor/pkg/api/plugins"
	"github.com/intel/edge-conductor/pkg/eputils"
)

func DockerCreate(in_config *api.ContainersItems0) (string, error) {
	networkMode := "bridge"
	if in_config.HostNetwork {
		networkMode = "host"
	}
	networkNames := []string{}
	mounts := []mount.Mount{}        //in_config.VolumeMounts
	volumes := map[string]struct{}{} //volumes for containers
	binds := []string{}              //in_config.BindMounts
	ports := []string{}              //in_config.Ports
	env := []string{}                //in_config.Env

	needimagepull := true
	if in_config.ImagePullPolicy == "Never" {
		needimagepull = false
	}
	// TODO: check if image exists then determine whether needimagepull if set "IfNotPresent"

	userInfo := in_config.UserInContainer
	if len(userInfo) <= 0 {
		currentUser, err := user.Current()
		if err != nil {
			log.Errorln("Failed to get current user information.", err)
			return "", err
		}
		userInfo = fmt.Sprintf("%s:%s", currentUser.Uid, currentUser.Gid)
	} else if userInfo == "auto" {
		userInfo = ""
	}

	for _, vol := range in_config.VolumeMounts {
		mounts = append(mounts,
			mount.Mount{
				Type:     mount.TypeVolume,
				Source:   vol.HostPath,
				Target:   vol.MountPath,
				ReadOnly: vol.ReadOnly,
			},
		)
	}

	for _, bind := range in_config.BindMounts {
		if bind.ReadOnly {
			binds = append(binds,
				//String format: /host/location:/container/location:type
				fmt.Sprintf("%s:%s:ro", bind.HostPath, bind.MountPath),
			)
		} else {
			binds = append(binds,
				//String format: /host/location:/container/location:type
				fmt.Sprintf("%s:%s", bind.HostPath, bind.MountPath),
			)
		}
	}

	for _, tmpfs := range in_config.Tmpfs {
		mounts = append(mounts,
			mount.Mount{
				Type:   mount.TypeTmpfs,
				Source: "",
				Target: tmpfs,
			},
		)
	}

	if len(in_config.Networks) != 0 {
		if networkMode == "host" {
			log.Warningf("HostNetwork is set to true, Networks parameter will be ignored.")
		} else {
			networkNames = append(networkNames, in_config.Networks...)
		}
	}

	for _, port := range in_config.Ports {
		protocol := "tcp"
		if len(port.Protocol) != 0 {
			protocol = port.Protocol
		}
		ip := port.HostIP
		if ip == "0.0.0.0" {
			return "", eputils.GetError("errIP")
		} else if ip == "" {
			// use "127.0.0.1" if ip is not specific
			ip = "127.0.0.1"
		} else if ip == "localhost" {
			ip = "127.0.0.1"
		}
		ports = append(ports,
			//String format: ip:public_port:private_port/proto
			fmt.Sprintf("%s:%d:%d/%s", ip, port.HostPort, port.ContainerPort, protocol),
		)
	}

	//TODO: Override env value
	for _, e := range in_config.Env {
		//String format: env_name=env_value
		env = append(env,
			fmt.Sprintf("%s=%s", e.Name, e.Value),
		)
	}

	readOnlyRootfs := true
	if in_config.ReadOnlyRootfs != nil {
		if !*in_config.ReadOnlyRootfs {
			readOnlyRootfs = false
			log.Infoln("Rootfs is not set as ReadOnly.")
		}
	}
	if readOnlyRootfs {
		log.Infoln("Rootfs is set as ReadOnly.")
	}

	restart := in_config.Restart

	ctnID, err := CreateContainer(
		// string
		in_config.Image,
		in_config.Name,
		in_config.HostName,
		networkMode,
		networkNames,
		userInfo,
		// bool
		false, // Privileged mode is not allowed.
		needimagepull,
		in_config.RunInBackground,
		readOnlyRootfs,
		// []string
		in_config.Command,
		in_config.Args,
		binds,
		mounts,
		volumes,
		ports,
		env,
		in_config.CapAdd,
		in_config.SecurityOpt,
		restart,
	)
	if err != nil {
		log.Errorln("Failed to create Container", in_config.Name)
		return "", err
	}
	return ctnID, nil
}

func DockerRun(in_config *api.ContainersItems0) error {
	if in_config.Force {
		container, err := GetContainerByName(in_config.Name)
		if err != nil {
			return err
		}

		if container != nil {
			// Container Exists
			if err := DockerRemove(in_config); err != nil {
				log.Errorln("Failed to remove", in_config.Name, err)
				return err
			}
		}
	}
	ctnID, err := DockerCreate(in_config)
	if err != nil {
		return err
	} else {
		if err = StartContainer(ctnID, in_config.Name, in_config.RunInBackground); err != nil {
			log.Errorln("Failed to start", ctnID, in_config.Name, err)
			return err
		}
	}
	return nil
}

func DockerStart(in_config *api.ContainersItems0) error {
	if err := StartContainer("", in_config.Name, in_config.RunInBackground); err != nil {
		log.Errorln("Failed to start", in_config.Name, err)
		return err
	}
	return nil
}

func DockerStop(in_config *api.ContainersItems0) error {
	if err := StopContainer(in_config.Name); err != nil {
		log.Errorln("Failed to stop", in_config.Name, err)
		return err
	}
	return nil
}

func DockerRemove(in_config *api.ContainersItems0) error {
	if err := RemoveContainer(in_config.Name); err != nil {
		log.Errorln("Failed to remove", in_config.Name, err)
		return err
	}
	return nil
}
