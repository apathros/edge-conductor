# How to enable K8S on CycloneV board and join it to a cluster

## Contents

  * [Configuring CycloneV Board](#configuring-cycloneV-board)
  * [Compiling Linux Kernel](#compiling-linux-kernel)
  * [Building Yocto image](#building--yocto-image)
  * [Creating SD Card image](#creating-sd-card-image)
  * [Adjusting time on the board](#adjusting-time-on-the-board)
  * [Installing and configuring docker service on the board](#installing-and-configuring-docker-service-on-the-board)
  * [Installing and configuring K8S services on the board](#installing-and-configuring-docker-services-on-the-board)
  * [Joining the board to a K8S cluster on the board](#joining-the-board-to-a-k8s-cluster-on-the-board)
&nbsp;

## Configuring CycloneV Board

This section presents the necessary board settings in order to run the GSRD on the Cyclone V SoC development board.

First, the board jumpers need to be configured as follows:
* J5: Open 
* J6: Short 
* J7: Short 
* J9: Open 
* J13: Short 
* J16: Open 
* J26: Short pins 1-2 
* J27: Short pins 2-3 
* J28: Short pins 1-2 
* J29: Short pins 2-3 
* J30: Short pins 1-2 
* J31: Open 

Then, the board switches need to be configured as follows:
* SW1: All OFF 
* SW2: All OFF 
* SW3: ON-OFF-ON-OFF-ON-ON 
* SW4: OFF-OFF-ON-ON 

&nbsp;

## Compiling Linux Kernel
```bash
mkdir cv_gsrd
cd cv_gsrd
export TOP_FOLDER=$(pwd)
cd $TOP_FOLDER
wget https://releases.linaro.org/components/toolchain/binaries/latest-7/arm-eabi/gcc-linaro-7.5.0-2019.12-x86_64_arm-eabi.tar.xz
tar xf gcc-linaro-7.5.0-2019.12-x86_64_arm-eabi.tar.xz
export PATH=$(pwd)/gcc-linaro-7.5.0-2019.12-x86_64_arm-eabi/bin:$PATH
export ARCH=arm
export CROSS_COMPILE=arm-eabi-


# Get the GSRD Linux Device Tree include file
wget https://releases.rocketboards.org/release/2020.07/gsrd/cv_gsrd/socfpga_cyclone5_ghrd.dtsi

# Get Linux source code
git clone https://github.com/altera-opensource/linux-socfpga
cd linux-socfpga
git checkout -b test ACDS20.1STD_REL_GSRD_PR

# Copy the GSRD DTS include file, and include it in the DTS
cp -f ../socfpga_cyclone5_ghrd.dtsi arch/arm/boot/dts
echo "#include \"socfpga_cyclone5_ghrd.dtsi\"" >> arch/arm/boot/dts/socfpga_cyclone5_socdk.dtsi

# Build kernel, device trees and modules
make socfpga_defconfig

#configuring Linux Kernel at this step
make menuconfig
make -j 24 zImage dtbs modules
make modules_install INSTALL_MOD_PATH=modules_install
rm -rf modules_install/lib/modules/*/build
rm -rf modules_install/lib/modules/*/source
```

The following files will be created:

File	Description

* Linux kernel image:    $TOP_FOLDER/linux-socfpga/arch/arm/boot/zImage	
* Linux Device Tree Blob:    $TOP_FOLDER/linux-socfpga/arch/arm/boot/dts/socfpga_cyclone5_socdk.dtb
* Folder with compiled loadable modules:    $TOP_FOLDER/linux-socfpga/modules_install


## Building Yocto image
```bash
cd $TOP_FOLDER
rm -rf yocto && mkdir yocto && cd yocto
git clone -b zeus https://git.yoctoproject.org/git/poky.git
git clone -b zeus https://git.openembedded.org/meta-openembedded
git clone -b zeus https://git.yoctoproject.org/git/meta-virtualization
git clone -b master https://github.com/kraj/meta-altera.git
git clone -b master https://github.com/altera-opensource/meta-altera-refdes.git
cd meta-altera-refdes
git checkout ACDS20.1STD_REL_GSRD_PR
```

*   Run script to initialize the build environment for bitbake
*   Note: You will be redirect to "build" folder once you execute the command below
```bash
cd ..
source poky/oe-init-build-env
```

*   Remove default settings imported from template
```bash
sed -i /MACHINE\ \?\?=\ \"qemux86-64\"/d conf/local.conf
```

*   Install K8S and Docker packages
```bash
echo 'DISTRO_FEATURES_append = " systemd"' >> conf/local.conf
echo 'VIRTUAL-RUNTIME_init_manager = "systemd"' >> conf/local.conf
echo 'CORE_IMAGE_EXTRA_INSTALL += " conntrack-tools ebtables socat util-linux iproute2 openvswitch"' >> conf/local.conf
```

*   Use the LTS kernel and set version
```bash
echo 'PREFERRED_PROVIDER_virtual/kernel = "linux-altera-lts"' >> conf/local.conf
echo 'PREFERRED_VERSION_linux-altera-lts = "5.4.23%"' >> conf/local.conf
```

*   Build additional rootfs type
```bash
echo 'IMAGE_FSTYPES += "tar.gz"' >> conf/local.conf
```

*   Settings for bblayers.conf
```bash
echo 'BBLAYERS += " ${TOPDIR}/../meta-altera "' >> conf/bblayers.conf
echo 'BBLAYERS += " ${TOPDIR}/../meta-altera-refdes "' >> conf/bblayers.conf
echo 'BBLAYERS += " ${TOPDIR}/../meta-openembedded/meta-oe "' >> conf/bblayers.conf
echo 'BBLAYERS += " ${TOPDIR}/../meta-openembedded/meta-networking "' >> conf/bblayers.conf
echo 'BBLAYERS += " ${TOPDIR}/../meta-openembedded/meta-python "' >> conf/bblayers.conf
echo 'BBLAYERS += " ${TOPDIR}/../meta-openembedded/meta-filesystems "' >> conf/bblayers.conf
echo 'BBLAYERS += " ${TOPDIR}/../meta-virtualization "' >> conf/bblayers.conf
```

*   Set the MACHINE, only at a time to avoid build conflict
```bash
echo "MACHINE = \"cyclone5\"" >> conf/local.conf
```

*   Ensure we build in all kernel-modules
```bash
echo "MACHINE_ESSENTIAL_EXTRA_RRECOMMENDS += \"kernel-modules\"" >> conf/local.conf
```

*   Build rootfs image
```bash
bitbake  gsrd-console-image
```

The following files will be created

File	Description

Compressed root filesystem archive:    $TOP_FOLDER/yocto/build/tmp/deploy/images/cyclone5/gsrd-console-image-cyclone5.tar.gz


## Creating SD card image
*   Create folder to keep the SD card binaries
```bash
cd $TOP_FOLDER
mkdir sd_card && cd sd_card
```

*   Get SD card creation script
```bash
wget https://releases.rocketboards.org/release/2020.05/gsrd/tools/make_sdimage_p3.py
chmod +x make_sdimage_p3.py
```

*   Get SPL image, refer to https://rocketboards.org/foswiki/Documentation/CycloneVSoCGSRD#Configuring_Serial_Connection
```bash
cp $TOP_FOLDER/cv_soc_devkit_ghrd/software/bootloader/u-boot-socfpga/u-boot-with-spl.sfp .
```

*   Prepare FAT partition contents, refer to https://rocketboards.org/foswiki/Documentation/CycloneVSoCGSRD#Configuring_Serial_Connection
```bash
mkdir fat && cd fat
cp $TOP_FOLDER/linux-socfpga/arch/arm/boot/zImage .
cp $TOP_FOLDER/linux-socfpga/arch/arm/boot/dts/socfpga_cyclone5_socdk.dtb .
cp $TOP_FOLDER/cv_soc_devkit_ghrd/output_files/*.rbf .
cp $TOP_FOLDER/u-boot.scr .
mkdir extlinux && cd extlinux
wget https://releases.rocketboards.org/release/2020.07/gsrd/cv_gsrd/extlinux.conf
cd ../..
```

*   Prepare Rootfs partition contents
```bash
mkdir rootfs && cd rootfs
sudo tar xf $TOP_FOLDER/yocto/build/tmp/deploy/images/cyclone5/gsrd-console-image-cyclone5.tar.gz
sudo rm -rf lib/modules/*
sudo cp -r $TOP_FOLDER/linux-socfpga/modules_install/lib/modules/* lib/modules/
cd ..
```

*   Build SD card image
```bash
sudo python3 make_sdimage_p3.py -f \
-P u-boot-with-spl.sfp,num=3,format=raw,size=10M,type=A2 \
-P rootfs/*,num=2,format=ext2,size=6000M \
-P fat/*,num=1,format=fat32,size=500M -s 8G \
-n sdcard.img 
```

## Powering on the board with SD card and adjusting time
```bash
date -s <time>
```

## Installing and configuring docker service on the board
```bash
#There is a problem of wget command in the board's OS, you can download the package on a host first, then copy the package from the host to the board
wget https://download.docker.com/linux/static/stable/armhf/docker-18.09.0.tgz
tar zxvf docker-18.09.0.tgz -C /usr/bin
chmod a+x /usr/bin/docker/*

touch /lib/systemd/system/docker.service
# docker.service contents as following
cat > /lib/systemd/system/docker.service << EOF
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
#BindsTo=containerd.service
After=network-online.target firewalld.service
Wants=network-online.target
#Requires=docker.socket

[Service]
Type=notify
# the default is not to use systemd for cgroups because the delegate issues still
# exists and systemd currently does not support the cgroup feature set required
# for containers run by docker
ExecStart=/usr/bin/docker/dockerd
ExecReload=/bin/kill -s HUP \$MAINPID
TimeoutSec=0
RestartSec=2
Restart=always

# Note that StartLimit* options were moved from "Service" to "Unit" in systemd 229.
# Both the old, and new location are accepted by systemd 229 and up, so using the old location
# to make them work for either version of systemd.
StartLimitBurst=3

# Note that StartLimitInterval was renamed to StartLimitIntervalSec in systemd 230.
# Both the old, and new name are accepted by systemd 230 and up, so using the old name to make
# this option work for either version of systemd.
StartLimitInterval=60s

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity

# Comment TasksMax if your systemd version does not support it.
# Only systemd 226 and above support this option.
TasksMax=infinity

# set delegate yes so that systemd does not reset the cgroups of docker containers
Delegate=yes

# kill only the docker process, not all processes in the cgroup
KillMode=process

[Install]
WantedBy=multi-user.target
EOF

# enable docker service
systemctl enable docker.service
systemctl daemon-reload && systemctl restart docker
```

## Installing and configuring K8S services on the board
```bash
#There is a problem of wget command in the board's OS, you can download the packages on a host first, then copy the packages from the host to the board
wget https://dl.k8s.io/v1.18.1/bin/linux/arm/kubelet
wget https://dl.k8s.io/v1.18.1/bin/linux/arm/kube-proxy
wget https://dl.k8s.io/v1.18.1/bin/linux/arm/kubeadm

cp kube* /usr/bin/
chmod a+x /usr/bin/kube*

# configure kubelet service
mkdir /etc/systemd/system/kubelet.service.d
cat > /etc/systemd/system/kubelet.service.d/10-kubeadm.conf << EOF
# contents of 10-kubeadm.conf
# Note: This dropin only works with kubeadm and kubelet v1.11+
[Service]
Environment="KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf"
Environment="KUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yaml"
# This is a file that "kubeadm init" and "kubeadm join" generates at runtime, populating the KUBELET_KUBEADM_ARGS variable dynamically
EnvironmentFile=-/var/lib/kubelet/kubeadm-flags.env
# This is a file that the user can use for overrides of the kubelet args as a last resort. Preferably, the user should use
# the .NodeRegistration.KubeletExtraArgs object in the configuration files instead. KUBELET_EXTRA_ARGS should be sourced from this file.
EnvironmentFile=-/etc/default/kubelet
ExecStart=
ExecStart=/usr/bin/kubelet \$KUBELET_KUBECONFIG_ARGS \$KUBELET_CONFIG_ARGS \$KUBELET_KUBEADM_ARGS \$KUBELET_EXTRA_ARGS
EOF

cat > /lib/systemd/system/kubelet.service << EOF
# contents of kubelet.service
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=https://kubernetes.io/docs/home/
Wants=network-online.target
After=network-online.target

[Service]
ExecStart=/usr/bin/kubelet
Restart=always
StartLimitInterval=0
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

cat > /lib/systemd/system/kube-proxy.service << EOF
# contents of kube-proxy.service
[Unit]
Description=Kubernetes Kube-Proxy Server
Documentation=https://kubernetes.io/docs/concepts/overview/components/#kube-proxy https://kubernetes.io/docs/reference/generated/kube-proxy/
After=network.target

[Service]
EnvironmentFile=-/etc/kubernetes/config
EnvironmentFile=-/etc/kubernetes/proxy
ExecStart=/usr/bin/kube-proxy \
            \$KUBE_LOGTOSTDERR \
            \$KUBE_LOG_LEVEL \
            \$KUBE_MASTER \
            \$KUBE_PROXY_ARGS
Restart=on-failure
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
```
## Joining the board to a K8S cluster on the board
*   join the cluster example:
```bash
kubeadm join <apiserver IP address>:6443 --token <token value>     --discovery-token-ca-cert-hash sha256:<CA cert hash value>
```


Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
