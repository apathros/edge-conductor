/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package docker

import (
	"context"
	api "ep/pkg/api/plugins"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

type DockerInterface interface {
	DockerCreate(in_config *api.ContainersItems0) (string, error)
	DockerRun(in_config *api.ContainersItems0) error
	DockerStart(in_config *api.ContainersItems0) error
	DockerStop(in_config *api.ContainersItems0) error
	DockerRemove(in_config *api.ContainersItems0) error
}

type DockerClientWrapperContainer interface {
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
	CreateContainer(
		imageName, containerName, hostName, networkMode string, networkNames []string,
		userInContainer string, privileged, needimagepull, runInBackground, readOnlyRootfs bool,
		entrypoint, args, binds []string, mounts []mount.Mount,
		volumes map[string]struct{},
		ports, env, capadd, securityOpt []string, restart string) (string, error)
	// StartContainer: Start a Container
	//   "docker start"
	//
	// Parameters:
	//   containerID:    A string of container ID.
	//   containerName:  A string of container Name.
	//   runInBackground: A bool value of whether run in background.
	// Note: If <containerID> is not provided, will query <containerName> for the ID.
	//
	StartContainer(containerID, containerName string, runInBackground bool) error
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
	RunContainer(
		imageName, containerName, hostName, networkMode string, networkNames []string,
		userInContainer string, privileged, needimagepull, runInBackground, readOnlyRootfs bool,
		entrypoint, args []string, binds []string,
		mounts []mount.Mount, volumes map[string]struct{},
		ports, env, capadd, securityOpt []string, restart string) (string, error)
	// StopContainer: Stop Container
	//   "docker stop <containerName>"
	//
	// Parameters:
	//   containerName: Name of the container to stop and remove.
	//
	StopContainer(containerName string) error
	// RemoveContainer: Stop and Remove Container
	//   "docker rm -f <containerName>"
	//
	// Parameters:
	//   containerName: Name of the container to stop and remove.
	//
	RemoveContainer(containerName string) error
	// GetContainerByName: Get a Container by its name
	//
	// Parameters:
	//   imageName: Name of the image to run.
	// Output:
	//   container: the pointer of a container object
	//
	GetContainerByName(containerName string) (*types.Container, error)
}

type DockerClientWrapperImage interface {
	// ImagePushToRegistry: Push to local registry
	//
	// Parameters:
	//   image:   Image to push
	//   registry: Registry path
	//   conf : Customconfig credential
	//
	ImagePushToRegistry(image string, registry string, conf *api.Customconfig) error
	// ImagePull: Pull image
	//
	// Parameters:
	//   imageRef:   Tag of the image
	//   authConf:   The authentication configuration
	//
	ImagePull(imageRef string, authConf *types.AuthConfig) error
	// ImagePush: Push image
	//
	// Parameters:
	//   imageRef:   Tag of the image
	//   authConf:   The authentication configuration
	//
	ImagePush(imageRef string, authConf *types.AuthConfig) error
	// ImageBuild: build image
	//
	// Parameters:
	//   dockerBuildTar: tar file to build image
	//   dockerFilePathInTar: Dockerfile relative path inside tar
	//   tag: Tag of the image to build
	//
	ImageBuild(dockerBuildTar string, dockerFilePathInTar string, tag string) error
	// ImageLoad: load image
	//
	// Parameters:
	//   tarball: tarball to load docker image
	//
	ImageLoad(tarball string) error
	// GetImageNewTag: Get a new tag of a image with the new registry URL.
	//
	// Parameters:
	//   imageTag:      Image Tag.
	//   registryURL:   New registry URL.
	// Output:
	//   newTag:        New Image Tag.
	//
	GetImageNewTag(imageTag, registryURL string) string
	// TagImageToLocal: Tag an image to a new tag use a new registry URL.
	//
	// Parameters:
	//   imageTag:      Image Tag.
	//   registryURL:   New registry URL.
	// Output:
	//   newTag:        New Image Tag.
	//
	TagImageToLocal(imageTag, registryURL string) (string, error)
	// TagImage: Tag an image to a new tag.
	//
	// Parameters:
	//   imageTag:      Image Tag.
	//   registryURL:   New image tag.
	//
	TagImage(imageTag, newTag string) error
	// TagImage: Get  Authconf from registry server
	//
	// Parameters:
	//   server: string.
	//   port: string
	//   user string
	//   password string
	//
	GetAuthConf(server, port, user, password string) (*types.AuthConfig, error)
	// GetHostImages: Get  images from Host
	//
	GetHostImages() (*map[string]int, error)
}

type DockerClientInterface interface {
	// ContainerList: github.com/docker/docker/client.ContainerList
	//
	ContainerList(cli *client.Client, ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	// ImagePull: github.com/docker/docker/client.ImagePull
	//
	ImagePull(cli *client.Client, ctx context.Context, refStr string, options types.ImagePullOptions) (io.ReadCloser, error)
	// ImagePush: github.com/docker/docker/client.ImagePush
	//
	ImagePush(cli *client.Client, ctx context.Context, image string, options types.ImagePushOptions) (io.ReadCloser, error)
	// ImageBuild: github.com/docker/docker/client.ImageBuild
	//
	ImageBuild(cli *client.Client, ctx context.Context, buildContext io.Reader, options types.ImageBuildOptions) (types.ImageBuildResponse, error)
	// ImageLoad: github.com/docker/docker/client.ImageLoad
	//
	ImageLoad(cli *client.Client, ctx context.Context, input io.Reader, quiet bool) (types.ImageLoadResponse, error)
	// ImageLoad: github.com/docker/docker/client.ImageList
	//
	ImageList(cli *client.Client, ctx context.Context, options types.ImageListOptions) ([]types.ImageSummary, error)
	// NetworkInspect: github.com/docker/docker/client.NetworkInspect
	//
	NetworkInspect(cli *client.Client, ctx context.Context, networkID string, options types.NetworkInspectOptions) (types.NetworkResource, error)
	// NetworkCreate: github.com/docker/docker/client.NetworkCreate
	//
	NetworkCreate(cli *client.Client, ctx context.Context, name string, options types.NetworkCreate) (types.NetworkCreateResponse, error)
	// ContainerCreate: github.com/docker/docker/client.ContainerCreate
	//
	ContainerCreate(cli *client.Client, ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, containerName string) (container.ContainerCreateCreatedBody, error)
	// ContainerRemove: github.com/docker/docker/client.ContainerRemove
	//
	ContainerRemove(cli *client.Client, ctx context.Context, containerID string, options types.ContainerRemoveOptions) error
	// ContainerStart: github.com/docker/docker/client.ContainerStart
	//
	ContainerStart(cli *client.Client, ctx context.Context, containerID string, options types.ContainerStartOptions) error
	// ContainerStop: github.com/docker/docker/client.ContainerStop
	//
	ContainerStop(cli *client.Client, ctx context.Context, containerID string, timeout *time.Duration) error
	// ContainerLogs: github.com/docker/docker/client.ContainerLogs
	//
	ContainerLogs(cli *client.Client, ctx context.Context, container string, options types.ContainerLogsOptions) (io.ReadCloser, error)
	// ContainerInspect: github.com/docker/docker/client.ContainerInspect
	//
	ContainerInspect(cli *client.Client, ctx context.Context, container string) (types.ContainerJSON, error)
	// ImageInspectWithRaw: github.com/docker/docker/client.ImageInspectWithRaw
	//
	ImageInspectWithRaw(cli *client.Client, ctx context.Context, imageID string) (types.ImageInspect, []byte, error)
	// ImageTag: github.com/docker/docker/client.ImageTag
	//
	ImageTag(cli *client.Client, ctx context.Context, source, target string) error
	// NewClientWithOpts: github.com/docker/docker/client.NewClientWithOpts
	//
	NewClientWithOpts(ops ...client.Opt) (*client.Client, error)
}
