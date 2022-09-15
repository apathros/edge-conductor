[Edge Conductor]: https://github.com/intel/edge-conductor
[Get Started]: ./get-started.md
[EC Configuration]: ./ec-configurations.md

# EC Error Code

This guide describes the way that Edge Conductor manages its error code.

## Error Code Structure

The Error Code in Edge Conductor includes three kinds of information:
- Error Code: alphanumeric ID of the error code. The first character of the error code must be the letters "E" and following with the numeric value presents the error in a different category. (refer to Table-1).

- Error Message: A string that contains information about the error. 

- Error Link: A URL pointing to the document which provides the possible solution to address the error. (Optional)

  

Table-1 Error Code Group 
| Error Code| Group            | Sub-Group|
|----| -----------------------------|---------|
| **E001** |`EC Tool Error`|N/A|
| E001.0xx |`EC Tool Error`| common config errors |
| E001.1xx |`EC Tool Error`| kind cluster errors  |
| E001.2xx |`EC Tool Error`| RKE cluster errors  |
| E001.3xx |`EC Tool Error`| CAPI error|
| E001.4xx |`EC Tool Error`|service errors|
| **E002** |`Network Error`| N/A |
| **E003** |`Kubernetes Error`| N/A |
| **E004** |`Security Error`| N/A |
| **E005** |`Utility Error`  | N/A |
| E005.0xx |`Utility Error`| Docker errors |
| E005.1xx |`Utility Error`| Harbor errors |
| E005.2xx |`Utility Error`| File utility errors |
| E005.3xx |`Utility Error`| Hash errors |
| E005.4xx |`Utility Error`| Repo utility errors |
| E005.5xx |`Utility Error`| ESP errors |

## Add an Error Code 
All error code definitions are in the [errorcode.go](../../../pkg/eputils/errorcode.go) file. Add a new error code by providing relevant information. Required fields are indicated with an asterisk.

ecode*: A unique alphanumeric ID that presents which group the error code belongs to. (refer to Table-1 for the error code group). If a new group is needed, follow the rule to extend the group or sub-group number. 

msg*: A string which contains information about the error. 

elink: if there is a document to provide the troubleshooting solution, put the URL of the document link here. If there is no specified document to provide, leave it empty.

```go
type EC_errors struct {
	ecode string
	msg   string
	elink string
}
```

Error Codes are collected in a map[] data structure with the index of the error key. 

Error key* :  A unique string as the index used by the other modules to get the ERROR type return value of the error code.

```go
var errorGroup = map[string]error{
	// E00: Unknown error
	"errUnknown": &EC_errors{"E000.099", "Unknown error", errorIndex},
	"errTest":    &EC_errors{"E000.100", "Test error", errorIndex},
	
	// E001 EC tools errors
	// E001.0xx common config errors
	"errKitConfig":  &EC_errors{"E001.001", "Edge Conductor Kit config file is not found", ""},
	"errConfigPath": &EC_errors{"E001.002", "Edge Conductor Kit config path is not correct", ""},
	"errParameter":  &EC_errors{"E001.003", "Edge Conductor Kit parameters is not found", ""},
	"errCustomCfg":  &EC_errors{"E001.004", "Edge Conductor Kit customconfig is not found", ""},
    ...
}
```


## Get the Error Code
From other modules, call the function 'GetError()` with the error key to retrieve the error.

For example:

```go
if kitconfig.Validate(nil) != nil {
		log.Warningf("Verify kitconfig err: %v", err)
		return eputils.GetError("errKitConfig")
	}
```




Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
