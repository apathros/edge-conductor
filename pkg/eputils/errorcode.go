/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

import (
	"fmt"
)

const (
	errorIndex = "https://github.com/intel/edge-conductor/tree/main/docs/troubleshooting/index.md"
)

type EC_errors struct {
	ecode string
	msg   string
	elink string
}

func (e *EC_errors) Error() string {
	if len(e.elink) == 0 {
		return fmt.Sprintf("%v: %v", e.ecode, e.msg)
	}
	return fmt.Sprintf("%v: %v (more information: %v)", e.ecode, e.msg, e.elink)
}

func (e *EC_errors) Code() string {
	return e.ecode
}

func (e *EC_errors) Msg() string {
	return e.msg
}

var ErrorGroup = map[string]error{
	// E000: Unknown error
	"errUnknown": &EC_errors{"E000.099", "Unknown error", errorIndex},
	"errTest":    &EC_errors{"E000.100", "test error", errorIndex},

	// E001: EC tools errors
	// E001.0**: common config errors
	"errKitConfig":  &EC_errors{"E001.001", "kitconfig section is not found in runtime file", ""},
	"errConfigPath": &EC_errors{"E001.002", "kitconfigpath section is not found in runtime file", ""},
	"errParameter":  &EC_errors{"E001.003", "Edge Conductor kit runtime file is not found ", ""},
	"errCustomCfg":  &EC_errors{"E001.004", "Edge Conductor kit customconfig is not found", ""},
	"errRegistryPw": &EC_errors{"E001.005", "Harbor Register password is not set", ""},
	"errCertCfg":    &EC_errors{"E001.006", "Cert manager config is missing in manifest", ""},
	"errYml":        &EC_errors{"E001.007", "Malformat Yaml file", ""},
	"errDockerCfg":  &EC_errors{"E001.008", "Can not find docker config folder", ""},
	"errSelector":   &EC_errors{"E001.009", "the 'name' section under components. selector is missing", ""},
	"errCompSlect":  &EC_errors{"E001.010", "component.selector section has error", ""},

	"errSelectorOverride": &EC_errors{"E001.011", "component.selector.override section has error", ""},
	"errUnmarshalOver":    &EC_errors{"E001.012", "'unmarshal' operation failed of the component.selector.override section", ""},
	"errMaxResource":      &EC_errors{"E001.013", "The length of Args exceeds the MAX COUNT(100)", ""},

	"errKitCfgParameter":        &EC_errors{"E001.014", "Edge Conductor Kit Config's parameter is not correct", ""},
	"errKitCfgParmMiss":         &EC_errors{"E001.015", "Edge Conductor Kit Config's parameter missing", ""},
	"errIncorrectParam":         &EC_errors{"E001.016", "Invalid parameter", ""},
	"errCopyFromDay0":           &EC_errors{"E001.017", "CopyFromDay0: must end with \"/\"", ""},
	"errCopyToDay0":             &EC_errors{"E001.018", "CopyToDay0: must end with \"/\"", ""},
	"errManifest":               &EC_errors{"E001.019", "cluster manifest is not found", ""},
	"errStringOverrideWithNode": &EC_errors{"E001.020", "StringOverrideWithNode", ""},
	"errResource":               &EC_errors{"E001.021", "resource is not specified in cluster manifest", ""},
	"errGrpcConnect":            &EC_errors{"E001.022", "grpc connect error", ""},
	"errWorkflow":               &EC_errors{"E001.023", "workflow is not found", ""},
	"errPluginComplete":         &EC_errors{"E001.024", "Plugin Complete error", ""},
	"errSchemaInitData":         &EC_errors{"E001.025", "Failed to get ep-params in init data", ""},
	"errCmdNotSupported":        &EC_errors{"E001.026", "This command is not supported for current config", ""},
	"errUnmarshalData":          &EC_errors{"E001.027", "Unmarshal data error", ""},
	"errUnmarshalPlugin":        &EC_errors{"E001.028", "Unmarshal plugin data error", ""},
	"errUnknownRet":             &EC_errors{"E001.029", "Unknown return value", ""},
	"errSchemaOutData":          &EC_errors{"E001.030", "Failed to get schema in output", ""},
	"errFind":                   &EC_errors{"E001.031", "Cannot find plugin", ""},
	"errPullFile":               &EC_errors{"E001.032", "Failed to find input schema", ""},
	"errInputSchema":            &EC_errors{"E001.033", "Cannot find input schema", ""},
	"errGetLogStream":           &EC_errors{"E001.034", "Get log stream error", ""},
	"errPluginConnect":          &EC_errors{"E001.035", "Plugin connection error", ""},
	"errConvert":                &EC_errors{"E001.036", "Failed to convert struct to map", ""},
	"errService":                &EC_errors{"E001.037", "Failed to convert map to service", ""},
	"errOverride":               &EC_errors{"E001.038", "Unmarshal override error", ""},
	"errLoadData":               &EC_errors{"E001.039", "failed to load ep-params data from init data", ""},
	"errPluginData":             &EC_errors{"E001.040", "failed to load plugin data from init data", ""},
	"errMarshalPdata":           &EC_errors{"E001.041", "marshal schemaMap data error", ""},
	"errPreviousSchema":         &EC_errors{"E001.042", "Failed to get schema in previous step's init or output data", ""},
	"errConvContainers":         &EC_errors{"E001.043", "convert containers data error", ""},
	"errMarshalInitData":        &EC_errors{"E001.044", "marshal initdata failed", ""},
	"errNodeLogin":              &EC_errors{"E001.045", "Node login username missing", ""},
	"errImage":                  &EC_errors{"E001.046", "image is not specified in cluster manifest", ""},
	"errNodeLoginPassword":      &EC_errors{"E001.047", "Node login password and ssh key missing", ""},
	"errIgnoreOverride":         &EC_errors{"E001.048", "Ignore override error", ""},
	"errIgnoreFormat":           &EC_errors{"E001.049", "Ignore format error", ""},
	"errUnknownCmdType":         &EC_errors{"E001.050", "Unknown command type", ""},
	"errBinary":                 &EC_errors{"E001.051", "binary is not specified in cluster manifest", ""},

	// E001.1**: kind cluster errors
	"errCreateKIND": &EC_errors{"E001.101", "Failed to create KIND cluster", ""},
	"errKINDImage":  &EC_errors{"E001.102", "Kind image info missing in manifest.", ""},
	"errDelKIND":    &EC_errors{"E001.103", "Failed to delete KIND cluster", ""},
	"errKindConfig": &EC_errors{"E001.104", "Kind info missing, pls check manifest", ""},

	// E001.2**: rke cluster errors
	"errRunRKE":    &EC_errors{"E001.202", "Failed to run rke command", ""},
	"errConfViper": &EC_errors{"E001.203", "Could not config viper.", ""},

	// E001.3**: CAPI Error
	"errCertMgrCfg":           &EC_errors{"E001.301", "Cert manager config is missing in manifest", ""},
	"errProvConfig":           &EC_errors{"E001.302", "CAPI provider parameter missing in manifest", ""},
	"errProvider":             &EC_errors{"E001.303", "require one infra provider, pls select one infra provider at kit config", ""},
	"errProviderLost":         &EC_errors{"E001.304", "CAPI providers missing in manifest, four providers are required ", ""},
	"errNumberNodes":          &EC_errors{"E001.305", "invalid worker node number pls check kit config", ""},
	"errCAPIManifest":         &EC_errors{"E001.306", "CAPI manifest missing", ""},
	"errCAPIInfraProvider":    &EC_errors{"E001.307", "validated CAPI infra provider missing in manifest", ""},
	"errCAPIKindLost":         &EC_errors{"E001.308", "kind config missing in manifest", ""},
	"errDownload":             &EC_errors{"E001.309", "capi download error", ""},
	"errClientGet":            &EC_errors{"E001.310", "error in getting clients", ""},
	"errInternalServer":       &EC_errors{"E001.311", "internal server error", ""},
	"errNoKindBinInReg":       &EC_errors{"E001.312", "Failed to find binary of kind in capi binary list", ""},
	"errAppendFile":           &EC_errors{"E001.320", "failed to generate capi binary list", ""},
	"errCAPIProvider":         &EC_errors{"E001.321", "failed to get capi clusterctl configuration", ""},
	"errGenProvRepoCctl":      &EC_errors{"E001.322", "failed to generate local provider repo for clusterctl", ""},
	"errGenCfgClusterctl":     &EC_errors{"E001.323", "failed to generate config files for clusterctl", ""},
	"errLaunchMgmtClster":     &EC_errors{"E001.324", "failed to launch management cluster", ""},
	"errInitClusterctl":       &EC_errors{"E001.325", "failed to init clusterapi for define target cluster", ""},
	"errStartMgmtClster":      &EC_errors{"E001.326", "failed to start/restart management cluster of CAPI", ""},
	"errRunClusterctlCmd":     &EC_errors{"E001.327", "Failed to run clusterctl cmd", ""},
	"errDeploymentLaunchFail": &EC_errors{"E001.328", "Failed to launch byoh controller manager", ""},
	"errNodeNotReady":         &EC_errors{"E001.329", "BYOH host not ready", ""},
	"errNode":                 &EC_errors{"E001.330", "no controller plane node ready", ""},
	"errMgmtCluster":          &EC_errors{"E001.331", "Failed to get management cluster binary list", ""},

	// E001.4**: Service errors
	"errExtNotFound":     &EC_errors{"E001.401", "service's tls extension of  is not found", ""},
	"errPullInvalidCmd":  &EC_errors{"E001.402", "pullFile: invalid command", ""},
	"errExtCfgNotFound":  &EC_errors{"E001.403", "service's tls config is not found", ""},
	"errCSRFileNotFound": &EC_errors{"E001.404", "Service TLS error: CSR filename not found", ""},
	"errPullOnlyOnDay0":  &EC_errors{"E001.405", "pullFile: only supported on day-0", ""},
	"errPushInvalidCmd":  &EC_errors{"E001.406", "pushFile: invalid command", ""},
	"errPushOnlyOnDay0":  &EC_errors{"E001.407", "pullFile: only supported on day-0", ""},
	"errHelmEmpty":       &EC_errors{"E001.408", "Helm repo or chart name is empty", ""},
	"errServiceStatus":   &EC_errors{"E001.409", "Helm service is in a wrong status", ""},
	"errUnknownStatus":   &EC_errors{"E001.410", "Helm service status is unknown", ""},
	"errPluginReturn":    &EC_errors{"E001.411", "failed to return schemaMap data", ""},
	"errNotInList":       &EC_errors{"E001.412", "File not found in download list.", ""},
	"errNoServerPort":    &EC_errors{"E001.413", "Server address or port is missing in kitconfig", ""},

	// E002: Network errors
	"errHost":           &EC_errors{"E002.002", " provide_ip under global setings is not set", ""},
	"errSSHPath":        &EC_errors{"E002.003", "SSH path for provision is not found.", ""},
	"errRemoteNotAFile": &EC_errors{"E002.004", "remote copy failed with invalid file path", ""},

	// E003: Kubernetes
	"errNok8sNode": &EC_errors{"E003.001", "No k8s node", ""},

	// E004: Security errors
	"errCertType":        &EC_errors{"E004.001", "unsupported certificate type", ""},
	"errCertDecodeFail":  &EC_errors{"E004.002", "failed to decode Cert", ""},
	"errCaCertIsDir":     &EC_errors{"E004.003", "certbundle.Ca.Cert is a directory", ""},
	"errCaKeyIsDir":      &EC_errors{"E004.004", "certbundle.Ca.Key is a directory", ""},
	"errServerCertIsDir": &EC_errors{"E004.005", "certbundle.Server.Cert is a directory", ""},
	"errServerKeyIsDir":  &EC_errors{"E004.006", "certbundle.Server.Key is a directory", ""},
	"errClientCertIsDir": &EC_errors{"E004.007", "certcertbundle.Client.Cert is a directory", ""},
	"errClientKeyIsDir":  &EC_errors{"E004.008", "certbundle.Client.Key is a directory", ""},
	"errCertNil":         &EC_errors{"E004.009", "cert path or Key path is nil", ""},
	"errKeyAlgo":         &EC_errors{"E004.010", "unsupported key algo", ""},
	"errRootCert":        &EC_errors{"E004.011", "failed to parse root certificate", ""},

	// E005: Utility errors
	// E005.0**: Docker errors
	"errNoAuth":              &EC_errors{"E005.001", "no auth in docker client configuration file", ""},
	"errNoRegistry":          &EC_errors{"E005.002", "registry server or port is not found", ""},
	"errInvalidPath":         &EC_errors{"E005.003", "invalid docker config path，check the DOCKER", ""},
	"errDockerClientConfig":  &EC_errors{"E005.004", "failed to open docker client config file", ""},
	"errNoContainer":         &EC_errors{"E005.005", "container name is not found", ""},
	"errNoDay0":              &EC_errors{"E005.006", "pushImage only supported on day-0", ""},
	"errPullingFile":         &EC_errors{"E005.007", "Pulling file failure.", ""},
	"errAbnormalExit":        &EC_errors{"E005.008", "docker start failed", ""},
	"errIP":                  &EC_errors{"E005.009", "host ip as \"0.0.0.0\"  in docker config  is not accepted", ""},
	"errParseCliCfg":         &EC_errors{"E005.010", "failed to parse docker config file", ""},
	"errDecode":              &EC_errors{"E005.011", "failed to decode authentication information in config file", ""},
	"errInvalidString":       &EC_errors{"E005.012", "auth string in docker config file is invalid", ""},
	"errDockerCltCfg":        &EC_errors{"E005.013", "docker config file is not existed", ""},
	"errFileEmpty":           &EC_errors{"E005.014", "File path should not be empty", ""},
	"errOrasDefaultResolver": &EC_errors{"E005.015", "Oras default resolver not found", ""},
	"errOrasResolver":        &EC_errors{"E005.016", "Oras resolver not found", ""},
	"errOras":                &EC_errors{"E005.017", "It's an oras error", ""},

	// E005.1**: Harbor errors
	"errHarborIPEmpty":  &EC_errors{"E005.101", "input harbor IP is empty", ""},
	"errHarborUser":     &EC_errors{"E005.102", "input harbor user is empty", ""},
	"errHarborPasswd":   &EC_errors{"E005.103", "input harbor password is empty", ""},
	"errInputAuthSrv":   &EC_errors{"E005.104", "input auth server address is empty", ""},
	"errHarborPort":     &EC_errors{"E005.105", "input harbor port is empty", ""},
	"errHarborUrlEmpty": &EC_errors{"E005.106", "input harbor url is empty", ""},
	"errHarborResponse": &EC_errors{"E005.107", "error in harbor response ", ""},
	"errHarborAbnormal": &EC_errors{"E005.108", "Abnormal harbor response received", ""},
	"errCertNull":       &EC_errors{"E005.109", "Cert file is null", ""},
	"errProjectName":    &EC_errors{"E005.110", "Harbor project name is empty", ""},
	"errAuthEmpty":      &EC_errors{"E005.111", "Harbor auth string is empty", ""},

	// E005.2**: File utility errors
	"errInvalidFile":    &EC_errors{"E005.201", "file is not valid", ""},
	"errTgzUncompress":  &EC_errors{"E005.202", "file is not valid tgz", ""},
	"errUnmarshal":      &EC_errors{"E005.203", "unmarshal operation failed", ""},
	"errMarshal":        &EC_errors{"E005.204", "marshal operation failed", ""},
	"errTarPath":        &EC_errors{"E005.205", "tar operation failed with invalid tar file path", ""},
	"errTgzHeader":      &EC_errors{"E005.206", "uncompress tgz failed with invalid file header", ""},
	"errFileExist":      &EC_errors{"E005.207", "create file failed, the file is existed", ""},
	"errLoadJson":       &EC_errors{"E005.208", "Failed to load json file", ""},
	"errInputArryEmpty": &EC_errors{"E005.209", "Files for cluster deployment are not found. Please run \"cluster build\" first", ""},
	"errUrlSchema":      &EC_errors{"E005.210", "URL Schema Not Supported", ""},
	"errNoFileDir":      &EC_errors{"E005.211", "no such file or directory", ""},

	// E005.3**: Hash errors
	"errShaCheckFailed": &EC_errors{"E005.301", "SHA256 check failed", ""},
	"errHash":           &EC_errors{"E005.302", "Hash check failed", ""},

	// E005.4**: Repo utility errors
	"errNoPushClient": &EC_errors{"E005.401", "push to repo failed", ""},
	"errNoPullClient": &EC_errors{"E005.402", "pull from repo failed", ""},

	// E005.5**: ESP errors
	"errOSSession":   &EC_errors{"E005.501", "Cannot find OS session in top config", ""},
	"errESPManifest": &EC_errors{"E005.502", "ESP manifest missing", ""},
	"errESPCodebase": &EC_errors{"E005.503", "Cannot find ESP codebase", ""},
	"errESPBuild":    &EC_errors{"E005.504", "Cannot find ESP build.sh script", ""},
	"errESPRun":      &EC_errors{"E005.505", "Cannot find ESP run.sh script", ""},
	"errOSProvider":  &EC_errors{"E005.506", "OS provider is wrong", ""},
}

func GetError(errName string) error {
	return ErrorGroup[errName]
}
