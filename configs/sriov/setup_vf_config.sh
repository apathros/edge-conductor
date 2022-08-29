#!/bin/bash
#
# Copyright (c) 2021 Intel Corporation.
#
# SPDX-License-Identifier: Apache-2.0
#

# This utility is used for creating VF of the NIC specifed by vendor/device id.

# $1 -  vendorId, like "0x8086", this is optional, or ""
# $2 -  deviceId, like "0x37d2 0x37d0"
# $3 -  sriov_numvfs, 0 means disable all VFs

VENDOR_PATH_LIST=""
DEV_PATH_LIST=""
NET_PF_NAME=""
NUM_OF_VF_CONFIG=0
NUM_OF_VFIO_VF_CONFIG=0
RCLOCAL_FILE="/etc/rc.local"
SRIOV_NUM_VFS_FN="sriov_numvfs"
SRIOV_MAX_VF_FN="sriov_totalvfs"
# Need to add supported vendors and devcies after verifying on it.
SUPPORTED_VENDOR=(0x8086)
# 0x37d2 is for X722 NIC
# 0x1521 is for I350 NIC
# 0x1563 is for X550 NIC
# Above supported NICs have been verified.
SUPPORTED_DEVICES=(0x37d2 0x37d0 0x1521 0x1563 0x1572 0x1574 0x1580 0x1581 0x1583 0x1584 0x1585 0x158a 0x158b 0x1592 0x1593 0x159b 0x15e4)

function usage {
    cat >&2 <<EOF
Usage:
$(basename $0)
Please Input 4 parameters as below:
    - VendorIds: specify the vendor id for the NIC used for SRIOV, like "0x8086 0x10ec"
    - DeviceIds: specify the device id for the NIC used for SRIOV, like "0x37d2 0x37d0"
    - PFName: specify the PF interface name for the NIC used for SRIOV, like "eno2"
    - NumOfVFs: specify the number of VFs that will be created for the NIC
    - NumOfVfioVFs: specify the number of VFs which will be detached from host driver
eg: ./sriov_vf_config.sh "0x8086" "0x37d2 0x37d0" 4 1
EOF
}

function enable_rclocal {
# create /etc/rc.local
    cat << EOF | sudo tee "${RCLOCAL_FILE}"
#!/bin/bash
EOF
    sudo chmod +x "${RCLOCAL_FILE}"
# create systemd service file
    cat << EOF | sudo tee "/etc/systemd/system/rc-local.service"
[Unit]
Description=/etc/rc.local Support
ConditionPathExists=/etc/rc.local

[Service]
ExecStart=/etc/rc.local start
TimeoutSec=0
StandardOutput=tty
RemainAfterExit=yes

[Install]
WantedBy=multi-user.target
EOF
    sudo systemctl enable rc-local
}

function enable_vf {
    declare -a dev_path_list=$1
    echo "${dev_path_list[@]}"

    # Get all nics that are not UP.
    declare -a nic_list
    if [ -z $NET_PF_NAME ]; then
        nic_list=($(grep -r '.*'  /sys/class/net/*/device/net/*/operstate | grep down))
        nic_list=(${nic_list[@]%net*})
        echo "NIC list for state DOWN:"
        echo "${nic_list[@]}"
    else
        netstate=($(grep -r '.*'  /sys/class/net/${NET_PF_NAME}/device/net/${NET_PF_NAME}/operstate | grep down))
        if [ ${netstate} = "down" ]; then
            nic_list=("/sys/class/net/${NET_PF_NAME}/device/")
            echo "NIC list for specific PF" ${NET_PF_NAME} ":"
            echo "${nic_list[@]}"
        else
            echo "The PF interface" ${NET_PF_NAME} "is not in the down state or doesn't exist."
            return 0
        fi
    fi

    # Filter NICs with state UP
    echo "NIC list for SRIOV VF creation:"
    declare -a sriov_list
    for item1 in ${dev_path_list[@]}; do
	item1=${item1%device*}
        for item2 in ${nic_list[@]}; do
            if [[ ${item1} = ${item2} ]]; then
                echo "$item1"
                sriov_list+=(${item1})
            fi
        done
    done

    echo "Enable VF for ${#sriov_list[*]} NICs"
    enable_rclocal
    for dev_path in ${sriov_list[@]}; do
        sriov_num_vf_file="${dev_path}${SRIOV_NUM_VFS_FN}"
	sriov_max_vf_file="${dev_path}${SRIOV_MAX_VF_FN}"

	if [[ ! -f ${sriov_max_vf_file} ]]; then
            echo "Not find sriov_totalvfs file"
        else
            sriov_max_vf=$(cat ${sriov_max_vf_file})
	    echo "${sriov_num_vf_file}"
	    echo "Max VF:${sriov_max_vf}"
	    if [[ ${NUM_OF_VF_CONFIG} -gt ${sriov_max_vf} ]]; then
                NUM_OF_VF_CONFIG=${sriov_max_vf}
		echo "We can only set VF num to ${sriov_max_vf}!"
            fi
        fi

	echo "0" > ${sriov_num_vf_file}
	echo "echo ${NUM_OF_VF_CONFIG} > ${sriov_num_vf_file}"
	echo ${NUM_OF_VF_CONFIG} > ${sriov_num_vf_file}
	# enable rc.local to make sure VF can be configured each time the node resets
        command_str="sudo sh -c \"echo ${NUM_OF_VF_CONFIG}  > ${sriov_num_vf_file}\""
        echo ${command_str} >> "${RCLOCAL_FILE}"
    done
}

function hostdev_detach_vf {
	declare -a nic_vf_list
	declare -a vf_id_list
	declare -a vf_driver

	vf_bdf_list=($(lspci -nn | grep "Virtual Function" | cut -d \  -f 1 | sed 's/^/0000:/g'))
	vf_id_list=($(lspci -nn | grep "Virtual Function" | sed "s/.*\(\[8086:[0-9]*[a-f]*[0-9]*[a-f]*[0-9]*[a-f]*[0-9]*[a-f]*\]\)/\1/" | cut -d ' ' -f1))
	vfio_vf_num=${NUM_OF_VFIO_VF_CONFIG}
	vfio_pci_folder="/sys/bus/pci/drivers/vfio-pci"

	if [ ${NUM_OF_VFIO_VF_CONFIG} -eq 0 ] || [ ! -d $vfio_pci_folder ]; then
		echo "No vfio-pci driver in this node or VF numuber of VM is 0, no VF is allocated for VM."
		return 0
	else
		echo "vfio-pci driver is set for ${NUM_OF_VFIO_VF_CONFIG} VFs"
	fi
	for bdf_item in ${vf_bdf_list[@]}; do
		vf_driver=($(lspci -s ${bdf_item} -v | grep driver | awk -F' ' '{print $NF}'))
		echo -n "${bdf_item}" > /sys/bus/pci/drivers/${vf_driver}/unbind
		vfio_vf_num=$((${vfio_vf_num}-1))
		if [ ${vfio_vf_num} -eq 0 ]; then
			break
		fi
	done

	id_item=($(echo ${vf_id_list} | sed 's/\[//g;s/\]//g;s/:/ /g'))
	echo "${id_item[@]}" > /sys/bus/pci/drivers/vfio-pci/new_id
	vfio_vf_num=$((${vfio_vf_num}-1))
	echo "${NUM_OF_VFIO_VF_CONFIG} VFs host driver is detached"
}

if [[ $# != 5 ]]; then
    usage
    exit 1
fi

declare -a vendors
declare -a devices

if [[ -z $1 ]]; then
    echo "No vendor Id specified!"
else
    vendors=(${1})
    for vendor in ${vendors[@]}; do
        if [[ "${SUPPORTED_VENDOR[@]}" =~ "${vendor}" ]]; then
            VENDOR_PATH_LIST=("${VENDOR_PATH_LIST[*]}" $(grep -r '.*'  /sys/class/net/*/device/vendor | grep ${vendor}))
        else
            echo "Vendor: ${vendor} is not supported!"
	    echo "Supported vendor list is below:"
	    echo ${SUPPORTED_VENDOR}
	    exit 0
        fi
    done
fi

if [[ -z $2 ]]; then
    echo "No device Id specified!"
    exit 0
else
    devices=(${2})
    for device in ${devices[@]}; do
        if [[ "${SUPPORTED_DEVICES[@]}" =~ "${device}" ]]; then
            DEV_PATH_LIST=("${DEV_PATH_LIST[*]}"  $(grep -r '.*'  /sys/class/net/*/device/device | grep ${device}))
        else
            echo "Device: ${device} is not supported!"
	    echo "Supported device id is below:"
	    echo "${SUPPORTED_DEVICES[@]}"
	    exit 0
        fi
    done
fi

if [[ -z $3 ]]; then
    echo "No PF name specified!"
else
    NET_PF_NAME=$3
fi

if [[ -z $4 ]]; then
    echo "No sriov_numvfs specified!"
    exit 0
else
    NUM_OF_VF_CONFIG=$4
fi

if [[ -z $5 ]]; then
    echo "No sriov vfio-pci VF specified!"
    exit 0
else
    NUM_OF_VFIO_VF_CONFIG=$5
fi

echo "========Start Creating VF for SRIOV NIC=========="
echo "NIC Info list:"
lspci -nn | grep Eth
echo "VENDOR_PATH_LIST:"
echo "${VENDOR_PATH_LIST[@]}"
echo "DEV_PATH_LIST:"
echo "${DEV_PATH_LIST[@]}"
echo "NUM_OF_VF_CONFIG: ${NUM_OF_VF_CONFIG}"

if [[ ! ${VENDOR_PATH_LIST[@]} ]]; then
# only consider deviceId
    if [[ ! -z $1 ]]; then
	echo "Could not find any matched NIC, please specify right vendor ID"
	exit 0
    fi

    if [[ -z "${DEV_PATH_LIST}" ]]; then
	echo "Could not find any device for id: ${devices[@]}"
	exit 0
    else
	echo "${DEV_PATH_LIST}"
	enable_vf "${DEV_PATH_LIST[*]}"
    fi
else
    if [[ ! ${DEV_PATH_LIST[@]} ]]; then
        echo "Could not find any device for id: ${devices[@]}"
        exit 0
    else
	declare -a RESULT
        for item1 in ${DEV_PATH_LIST[@]}; do
            for item2 in ${VENDOR_PATH_LIST[@]}; do
                if [[ ${item1%device*} = ${item2%vendor*} ]]; then
                    RESULT+=($item1)
		fi
            done
        done

	enable_vf "${RESULT[*]}"
    fi
fi

echo "========End Creating VF for SRIOV NIC============"
echo "Virtual Function List:"
lspci -nn | grep "Virtual Function"
modprobe vfio
modprobe vfio-pci
hostdev_detach_vf
