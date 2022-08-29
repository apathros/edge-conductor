/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

//go:generate mockgen -destination=./mock/cliapi_mock.go -package=mock -copyright_file=../../../api/schemas/license-header.txt ep/pkg/eputils/docker DockerClientWrapperContainer,DockerClientWrapperImage,DockerClientInterface

package docker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"ep/pkg/eputils"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	api "ep/pkg/api/plugins"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/pkg/jsonmessage"
	"github.com/moby/term"
	log "github.com/sirupsen/logrus"
)

const (
	MAXRETRYCOUNT   = 5
	NONEWPRIVILEGES = "no-new-privileges"
	REGISTRYPROJECT = "/library"
)

var keyList = []string{"http_proxy", "https_proxy", "no_proxy", "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"}
var gcli *client.Client

// getDefaultContext: Get the default context.
//
func getDefaultContext() context.Context {
	return context.Background()
}

// getDockerClient: Get the global docker client object.
//
func getDockerClient() (*client.Client, error) {
	if gcli == nil {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Errorln(err)
			return nil, err
		}
		gcli = cli
	}
	return gcli, nil
}

// GetContainerByName: Get a Container by its name
//
// Parameters:
//   imageName: Name of the image to run.
// Output:
//   container: the pointer of a container object
//
func GetContainerByName(containerName string) (*types.Container, error) {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}

	accurateFilterName := "^/" + containerName + "$"
	filter := filters.NewArgs(filters.KeyValuePair{Key: "name", Value: accurateFilterName})

	containers, err := cli.ContainerList(
		ctx,
		types.ContainerListOptions{
			All:     true,
			Filters: filter})
	if err != nil {
		log.Errorln("ERROR: Failed to find", containerName, err)
		return nil, err
	}
	if len(containers) > 0 {
		return &containers[0], nil
	}
	return nil, nil
}

// ImagePushToRegistry: Push to local registry
//
// Parameters:
//   image:   Image to push
//   registry: Registry path
//   conf : Customconfig credential
//
func ImagePushToRegistry(image string, registry string, conf *api.Customconfig) error {
	log.Debugf("Push %s to %s/%s", image, registry, image)

	registryheadwithproj := registry + REGISTRYPROJECT

	registryusr := conf.Registry.User
	registrypasswd := conf.Registry.Password

	auth := types.AuthConfig{
		Username:      registryusr,
		Password:      registrypasswd,
		ServerAddress: registry,
	}

	newTag, err := TagImageToLocal(image, registryheadwithproj)
	if err != nil {
		return err
	}

	if err := ImagePush(newTag, &auth); err != nil {
		return err
	}

	return nil
}

// ImagePull: Pull image
//
// Parameters:
//   imageRef:   Tag of the image
//   authConf:   The authentication configuration
//
func ImagePull(imageRef string, authConf *types.AuthConfig) error {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	if authConf == nil {
		if authDefault, err := LoadDockerCliCredentials(imageRef); err != nil {
			return err
		} else {
			authConf = authDefault
		}
	}

	var authStr string
	if authConf != nil {
		encodedJSON, err := json.Marshal(authConf)
		if err != nil {
			return err
		}
		authStr = base64.URLEncoding.EncodeToString(encodedJSON)
	} else {
		authStr = ""
	}

	logreader, err := cli.ImagePull(ctx, imageRef, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		log.Errorf("Failed to pull image %s: %s", imageRef, err)
		return err
	}
	defer logreader.Close()

	terminalFD, isTerminal := term.GetFdInfo(os.Stdout)
	if err = jsonmessage.DisplayJSONMessagesStream(logreader, os.Stdout, terminalFD, isTerminal, nil); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// ImagePush: Push image
//
// Parameters:
//   imageRef:   Tag of the image
//   authConf:   The authentication configuration
//
func ImagePush(imageRef string, authConf *types.AuthConfig) error {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	var authStr string
	if authConf != nil {
		encodedJSON, err := json.Marshal(authConf)
		if err != nil {
			return err
		}
		authStr = base64.URLEncoding.EncodeToString(encodedJSON)
	} else {
		authStr = ""
	}

	logreader, err := cli.ImagePush(ctx, imageRef, types.ImagePushOptions{RegistryAuth: authStr})
	if err != nil {
		log.Errorf("Failed to push image %s", imageRef)
		return err
	}
	defer logreader.Close()

	terminalFD, isTerminal := term.GetFdInfo(os.Stdout)
	if err = jsonmessage.DisplayJSONMessagesStream(logreader, os.Stdout, terminalFD, isTerminal, nil); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// ImageBuild: build image
//
// Parameters:
//   dockerBuildTar: tar file to build image
//   dockerFilePathInTar: Dockerfile relative path inside tar
//   tag: Tag of the image to build
//
func ImageBuild(dockerBuildTar, dockerFilePathInTar, tag string) error {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	valid := eputils.IsValidFile(dockerBuildTar)
	if !valid {
		return eputils.GetError("errInvalidFile")
	}

	dockerBuildContext, err := os.Open(dockerBuildTar)
	if err != nil {
		return err
	}

	defer dockerBuildContext.Close()

	buildArgs := make(map[string]*string)

	for _, key := range keyList {
		if os.Getenv(key) != "" {
			buildArgs[key] = func() *string { v := os.Getenv(key); return &v }()
		}
	}

	buildResponse, err := cli.ImageBuild(ctx, dockerBuildContext, types.ImageBuildOptions{
		BuildArgs:  buildArgs,
		Remove:     true,
		Tags:       []string{tag},
		Dockerfile: dockerFilePathInTar,
	})
	if err != nil {
		log.Errorf("Failed to build image %s", err)
		return err
	}

	response, err := ioutil.ReadAll(buildResponse.Body)
	log.Infof("ImageBuild: response %s", response)

	if errClose := buildResponse.Body.Close(); errClose != nil {
		log.Error(errClose)
		return errClose
	}

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// ImageLoad: load image
//
// Parameters:
//   tarball: tarball to load docker image
//
func ImageLoad(tarball string) error {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	_, err = os.Stat(tarball)
	if err != nil {
		log.Errorf("No tarball %s", tarball)
		return err
	}

	valid := eputils.IsValidFile(tarball)
	if !valid {
		return eputils.GetError("errInvalidFile")
	}

	dockerLoadContext, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer dockerLoadContext.Close()

	loadResponse, err := cli.ImageLoad(ctx, dockerLoadContext, true)
	if err != nil {
		log.Errorf("Failed to load image %s", err)
		return err
	}

	response, err := ioutil.ReadAll(loadResponse.Body)
	log.Infof("ImageLoad: response %s", response)

	if errClose := loadResponse.Body.Close(); errClose != nil {
		log.Error(errClose)
		return errClose
	}

	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

// CreateContainer: Create a Container
//   "docker create"
//
// Parameters:
//   imageName:      Name of the image to run.
//   containerName:  Name of the container to create.
//   hostName:       Hostname of the container.
//   networkMode:    Run the container with Network Mode.
//   networkNames:   Bridge names of the container runs on, by default "bridge".
//   userInContainer: User that will run inside the container, also support user:group.
//   privileged:     If run the container with privileged mode.
//   needimagepull:  If pull image before run.
//   runInBackground:   If detach the container and run in background.
//   readOnlyRootfs: If rootfs request readonly
//   entrypoint:     A list of string of cmd entrypoint to run in container.
//   args:           A list of arguments of the cmd entrypoint.
//   binds:   A list of strings for volumn bindings.
//            String format: /host/location:/container/location
//   mounts:  Mounts specs used by the container
//   volumes: A list of strings for volume mount.
//   ports:   A list of strings for network port bindings.
//            String format: ip:public_port:private_port/proto
//   env:     A list of strings for environment variables.
//            String format: env_name=env_value
//   capadd:  A list of strings for CapAdd.
//   securityOpt:  A list of strings for Security Options.
//   restart: Restart policy to be used for the container
//
// Output:
//   containerID:    A string of container ID.
//
func CreateContainer(
	imageName, containerName, hostName, networkMode string, networkNames []string,
	userInContainer string, privileged, needimagepull, runInBackground, readOnlyRootfs bool,
	entrypoint, args, binds []string, mounts []mount.Mount,
	volumes map[string]struct{},
	ports, env, capadd, securityOpt []string, restart string) (string, error) {

	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return "", err
	}

	if needimagepull {
		err := ImagePull(imageName, nil)
		if err != nil {
			return "", err
		}
	}
	// Add NONEWPRIVILEGES to securityOpt
	securityOpt = append(securityOpt, NONEWPRIVILEGES)

	resources := container.Resources{
		PidsLimit: nil,
		Ulimits:   nil,
	}

	restartPolicy := container.RestartPolicy{
		Name:              restart,
		MaximumRetryCount: MAXRETRYCOUNT,
	}

	var host_config *container.HostConfig = &container.HostConfig{
		Privileged:     privileged,
		CapAdd:         strslice.StrSlice(capadd),
		AutoRemove:     false,
		NetworkMode:    container.NetworkMode(networkMode),
		Binds:          binds,
		RestartPolicy:  restartPolicy,
		ReadonlyRootfs: readOnlyRootfs,
		SecurityOpt:    securityOpt,
		Resources:      resources,
		Mounts:         mounts}

	net_config := &network.NetworkingConfig{
		EndpointsConfig: make(map[string]*network.EndpointSettings),
	}
	exposedPorts := make(nat.PortSet)
	var portBindings map[nat.Port][]nat.PortBinding

	if networkMode != "host" {
		exposedPorts, portBindings, err = nat.ParsePortSpecs(ports)
		if err != nil {
			log.Errorln("Failed to parse ports", ports, err)
			return "", err
		}
		host_config.PortBindings = portBindings

		if len(networkNames) != 0 {
			for _, net := range networkNames {
				_, err := cli.NetworkInspect(ctx, net,
					types.NetworkInspectOptions{Scope: "local", Verbose: false})
				if err != nil {
					_, err := cli.NetworkCreate(ctx, net, types.NetworkCreate{})
					if err != nil {
						log.Errorf("Create network error: %s", err)
						return "", err
					}
				}
				net_config.EndpointsConfig[net] = &network.EndpointSettings{}
			}
		}
	}

	resp, err := cli.ContainerCreate(
		// Context
		ctx,
		// Container Config
		&container.Config{
			Hostname:     hostName,
			Image:        imageName,
			Entrypoint:   entrypoint[:],
			Cmd:          args[:],
			Env:          env,
			User:         userInContainer,
			AttachStdin:  !runInBackground,
			AttachStdout: !runInBackground,
			AttachStderr: !runInBackground,
			StdinOnce:    !runInBackground,
			Tty:          false,
			Volumes:      volumes,
			ExposedPorts: exposedPorts,
		},
		// Host Config
		host_config,
		// Network Config
		net_config,
		// Container Name
		containerName)

	if err != nil {
		log.Errorln("Failed to create container", containerName, err)
		return "", err
	}

	log.Infoln("Container", resp.ID, containerName, "Successfully Created.")
	return resp.ID, nil
}

// StartContainer: Start a Container
//   "docker start"
//
// Parameters:
//   containerID:    A string of container ID.
//   containerName:  A string of container Name.
//   runInBackground: A bool value of whether run in background.
// Note: If <containerID> is not provided, will query <containerName> for the ID.
//
func StartContainer(containerID, containerName string, runInBackground bool) error {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	ID := containerID

	if len(ID) <= 0 {
		container, err := GetContainerByName(containerName)
		if err != nil {
			log.Errorln(err)
			return err
		}
		if container != nil {
			ID = container.ID
		} else {
			log.Errorln("Failed to find", containerName)
			return eputils.GetError("errNoContainer")
		}
	}

	// Trap Ctrl-C when foreground container is running
	// Remove the container and exit the process
	if !runInBackground {
		intCh := make(chan os.Signal, 1)
		signal.Notify(intCh, os.Interrupt, syscall.SIGINT)
		defer signal.Reset(os.Interrupt)
		finishCh := make(chan int)
		defer close(finishCh)
		go func() {
			select {
			case <-intCh:
				if err := cli.ContainerRemove(ctx, ID,
					types.ContainerRemoveOptions{
						RemoveVolumes: true,
						Force:         true,
					}); err != nil {
					log.Errorln("ERROR: Failed to remove", containerName, err)
				}
				os.Exit(1)
			case <-finishCh:
			}
		}()
	}

	if err := cli.ContainerStart(ctx, ID, types.ContainerStartOptions{}); err != nil {
		log.Errorln("Failed to start container", ID, err)
		return err
	}
	log.Infoln("Container", ID, containerName, "Successfully Started.")

	// Wait for container finished in foreground mode
	if !runInBackground {
		options := types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Timestamps: true,
			Follow:     true,
			Details:    true,
		}

		log.Infoln("Trying to get logs of", ID, containerName)

		logreader, err := cli.ContainerLogs(ctx, ID, options)
		if err != nil {
			log.Errorf("Failed to get log of container %s", ID)
			return err
		}
		defer logreader.Close()

		_, err = stdcopy.StdCopy(os.Stdout, os.Stderr, logreader)
		if err != nil {
			log.Errorf("Failed to Copy logreader : %s", err)
		}

		containerInfo, err := cli.ContainerInspect(ctx, ID)
		if err != nil {
			log.Errorf("Failed to get informations of container %s", ID)
			return err
		}
		if reflect.DeepEqual(containerInfo, types.ContainerJSON{}) {
			return eputils.GetError("errAbnormalExit")
		} else {
			if containerInfo.ContainerJSONBase.State.ExitCode != 0 {
				log.Errorf("Container exit code %d", containerInfo.ContainerJSONBase.State.ExitCode)
				return eputils.GetError("errAbnormalExit")
			}
		}
	}
	return nil
}

// RunContainer: Run Container at Background
//   "docker run"
//
// Parameters:
//   imageName:      Name of the image to run.
//   containerName:  Name of the container to create.
//   hostName:       Hostname of the container.
//   networkMode:    Run the container with Network Mode.
//   userInContainer: User that will run inside the container, also support user:group.
//   privileged:     If run the container with privileged mode.
//   needimagepull:  If pull image before run.
//   runInBackground:   If detach the container and run in background.
//   readOnlyRootfs: If rootfs request readonly
//   entrypoint:     A list of string of cmd entrypoint to run in container.
//   args:           A list of arguments of the cmd entrypoint.
//   binds:   A list of strings for volumn bindings.
//            String format: /host/location:/container/location
//   mounts:  Mounts specs used by the container
//   volumes: A list of strings for volume mount.
//   ports:   A list of strings for network port bindings.
//            String format: ip:public_port:private_port/proto
//   env:     A list of strings for environment variables.
//            String format: env_name=env_value
//   capadd:  A list of strings for CapAdd.
//   securityOpt:  A list of strings for Security Options.
//   restart: Restart policy to be used for the container
//
// Output:
//   containerID:    A string of container ID.
//
func RunContainer(
	imageName, containerName, hostName, networkMode string, networkNames []string,
	userInContainer string, privileged, needimagepull, runInBackground, readOnlyRootfs bool,
	entrypoint, args []string, binds []string,
	mounts []mount.Mount, volumes map[string]struct{},
	ports, env, capadd, securityOpt []string, restart string) (string, error) {

	ctnID, err := CreateContainer(
		imageName, containerName, hostName, networkMode, networkNames, userInContainer,
		privileged, needimagepull, runInBackground, readOnlyRootfs,
		entrypoint, args, binds, mounts, volumes,
		ports, env, capadd, securityOpt, restart,
	)
	if err != nil {
		log.Errorln("Failed to create container with", imageName)
		return "", err
	}

	if err := StartContainer(ctnID, containerName, runInBackground); err != nil {
		log.Errorln("Failed to start container", containerName, err)
		return ctnID, err
	}
	return ctnID, nil
}

// StopContainer: Stop Container
//   "docker stop <containerName>"
//
// Parameters:
//   containerName: Name of the container to stop and remove.
//
func StopContainer(containerName string) error {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	container, err := GetContainerByName(containerName)
	if err != nil {
		log.Errorln(err)
		return err
	}
	if container != nil && container.State == "running" {
		log.Infoln("Stopping container ", container.ID[:10], "... ")
		if err := cli.ContainerStop(ctx, container.ID, nil); err != nil {
			log.Errorln("Failed to stop", containerName, err)
			return err
		}
		log.Infoln("Successfully stopped", containerName)
	}

	return nil
}

// RemoveContainer: Stop and Remove Container
//   "docker rm -f <containerName>"
//
// Parameters:
//   containerName: Name of the container to stop and remove.
//
func RemoveContainer(containerName string) error {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	err = StopContainer(containerName)
	if err != nil {
		return err
	}

	container, err := GetContainerByName(containerName)
	if err != nil {
		log.Errorln(err)
		return err
	}
	if container != nil {
		log.Infoln("Removing container", container.ID[:10], "... ")
		if err := cli.ContainerRemove(
			ctx, container.ID,
			types.ContainerRemoveOptions{
				RemoveVolumes: true,
				Force:         true,
			}); err != nil {
			log.Errorln("ERROR: Failed to remove", containerName, err)
			return err
		}
		log.Infoln("Successfully removed", containerName)
	}

	return nil
}

// GetImageNewTag: Get a new tag of a image with the new registry URL.
//
// Parameters:
//   imageTag:      Image Tag.
//   registryURL:   New registry URL.
// Output:
//   newTag:        New Image Tag.
//
func GetImageNewTag(imageTag, registryURL string) string {
	imgName := imageTag

	// Remove digest reference in the name
	if strings.Contains(imgName, "@") {
		imgName = strings.Split(imgName, "@")[0]
	}

	// Remove registry url in the name
	nameSplit := strings.Split(imgName, "/")
	if len(nameSplit) > 1 && strings.Contains(nameSplit[0], ":") {
		imgName = strings.Replace(imgName, nameSplit[0]+"/", "", 1)
	}
	if len(imgName) > 0 {
		imgName = registryURL + "/" + imgName
	}

	return imgName
}

// TagImageToLocal: Tag an image to a new tag use a new registry URL.
//
// Parameters:
//   imageTag:      Image Tag.
//   registryURL:   New registry URL.
// Output:
//   newTag:        New Image Tag.
//
func TagImageToLocal(imageTag, registryURL string) (string, error) {
	ctx := getDefaultContext()
	cli, err := getDockerClient()
	if err != nil {
		return "", err
	}

	imginspect, _, err := cli.ImageInspectWithRaw(ctx, imageTag)
	if err != nil {
		log.Errorln("Failed to inspect Docker image:", imageTag)
		return "", err
	}

	ID := imginspect.ID
	newTag := GetImageNewTag(imageTag, registryURL)

	log.Infof("Tagging image %s to %s", ID, newTag)
	if err = cli.ImageTag(ctx, ID, newTag); err != nil {
		log.Error(err)
		return "", err
	}

	return newTag, nil
}

// TagImage: Tag an image to a new tag.
//
// Parameters:
//   imageTag:      Image Tag.
//   registryURL:   New image tag.
//
func TagImage(imageTag, newTag string) error {
	ctx := getDefaultContext()

	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	imginspect, _, err := cli.ImageInspectWithRaw(ctx, imageTag)
	if err != nil {
		log.Errorln("Failed to inspect Docker image:", imageTag)
		return err
	}

	ID := imginspect.ID
	err = cli.ImageTag(ctx, ID, newTag)
	if err != nil {
		log.Errorf("Failed to tag image %s to %s", ID, newTag)
		return err
	}

	return nil
}

// GetHostImages: Get All images from Host.
//
func GetHostImages() (*(map[string]int), error) {
	ctx := getDefaultContext()

	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}

	hostImageList, err := cli.ImageList(ctx, types.ImageListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	imageInject := make(map[string]int)
	for _, image := range hostImageList {
		for i, val := range image.RepoTags {
			imageInject[val] = i
		}
	}

	return &imageInject, nil
}
