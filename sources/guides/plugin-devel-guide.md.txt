# Plugin Development Guide

This guide describes how to develop Edge Conductor plugins, including an overview and some simple examples that help you use plugins and workflows.

- [Development Process Overview](#development-process-overview)
    - [Flow Diagram](#flow-diagram)
- [Coding Language and Standard](#coding-language-and-standard)
- [Edge Conductor Commands vs. Workflows](#edge-conductor-commands-vs-workflows)
- [Development Examples](#development-examples)
    - [Example 1: Generate a Simple Hello-World Plugin](#example-1-generate-a-simple-hello-world-plugin)
    - [Example 2: Generate Plugins with Input and Output Schemas](#example-2-generate-plugins-with-input-and-output-schemas)
    - [Example 3: Connect Plugins into an Existing Workflow](#example-3-connect-plugins-into-an-existing-workflow)

## Development Process Overview

Prerequisite: Gather the requirements of the new feature.

1.  Evaluate the plugin requirements.
    *   Divide the implementation of a WF into steps.
    *   Check if a step can be implemented with an existing plugin.
    *   Define the new plugins for other steps.
2.  Define the plugin relationships and define the interfaces of the new plugins.
    *   Add data schema details under the `api/schemas/plugins/` directory.
        *   Only schema files under `api/schemas/plugins` can be used as input/output by plugins, if a plugin needs schema files under other folders, please create a soft-link under `api/schemas/plugins`.
    *   Modify `pkg/epplugins/plugins.yml` with input/output details described with data schemas.
        *   Follow `data schema standard` in [Coding Language and Standard](#coding-language-and-standard) to write data schema files under `api/schemas` folder.
        
    *   Modify the file of a specific WF under `configs/workflow/` directory to insert the new plugins and connect the plugins in WFs. `configs/workflow/common/init.yml` will be used as an example in this guide.
3.  Run `make` to generate the skeleton code.
4.  Implement `PluginMain()` in the `pkg/epplugs/<plugin name>/main.go`
    directory.
5. Run `make` again. The plugin will be ready now.

### Flow Diagram
```mermaid
flowchart TD
st(Requirement) --> requirement[Define plugins based on the requirement]
requirement --> con{Need data schema?}
con --> |Yes| define_schema[Define the plugin input&#47output and data schema]
define_schema -. create .-> schema_yaml>api&#47schemas&#47plugins&#47&#60schema name&#62.yml]
con --> |No| define_plugin[Define the plugin relationships and interfaces]
define_plugin -.modify .-> plugins>pkg&#47epplugins&#47plugins.yml]
define_schema --> define_plugin
define_plugin --> insert_WF[Insert plugins into WF]
insert_WF -.modify .-> workflow>configs&#47workflow&#47common&#47&#60WF name&#62.yml]
insert_WF --> make[code gen tool &#40make&#41]
make -. create .->ep_plugins>pkg&#47epplugins&#47&#60plugin name&#62&#47*.go]
make --> implement_function(implement PluginMain&#40&#41)
implement_function -. modify .-> ep_plugins
implement_function --> make_again[make]
```

## Coding Language and Standard

*   Golang: [How to Write Go Code](https://golang.org/doc/code)
*   Go template for yaml files: [Go Template](https://pkg.go.dev/text/template)
    *   All yaml files used in the tool are following the template format defined by Golang.
    *   The Sprig library provides functions for Go’s template language: [Sprig Function Documentation](http://masterminds.github.io/sprig/)
    *   See an example: [Example of Yaml Template Used in the Project](../../configs/workflow/init/harbor.yml)
*   Data schema standard: [Schema Generation Rules](https://goswagger.io/use/models/schemas.html)
    *   Currently the following types are supported:
        - string
        - boolean
        - integer
        - object
        - array


## Edge Conductor Commands vs. Workflows

Current Edge Conductor commands are mapped to workflows as the following form:

| Commands  | Subcommand | Workflow Name     | Description |
| --------- | ---------- | ----------------- | ----------- |
| init      | -          | init              | Initialize Edge Conductor runtime config and environment. |
| deinit    | -          | deinit            | Clean Edge Conductor runtime environment. |
| cluster   | build      | cluster-build     | Build cluster configurations. |
| cluster   | deploy     | cluster-deploy    | Deploy cluster. |
| cluster   | reconcile  | cluster-reconcile | Reconcile cluster and generate kubeconfig. |
| cluster   | join       | node-join         | Join new nodes to an existing cluster. |
| service   | build      | service-build     | Build service configurations. |
| service   | deploy     | service-deploy    | Deploy services to the cluster. |
| service   | list       | service-list      | List current services with deploy status. |

## Development Examples

This section contains several examples of how to generate a plugin and insert
it into a workflow.

First, prepare the Edge Conductor codebase according to the [README](../../README.md).

### Example 1: Generate a Simple Hello-World Plugin

Let's start with a simple plugin, which prints a "hello-world" message. It does
not have any input or output schema for data transferring.

1.  Edit `pkg/epplugins/plugins.yml` to define the new plugin:

    ```
    - name: hello-world
    ```

2.  `make` the Edge Conductor codebase. The code for the `hello-world` plugin
    will be auto-generated:

    ```
    pkg/epplugins/hello-world/
    ├── generated.go
    ├── generated_testutil.go
    ├── main.go
    └── main_test.go
    ```

3.  Add functional logic in the entry `PluginMain` in
`pkg/epplugins/hello-world/main.go`:

    ```
    func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
            log.Infof(" >>> Plugin: hello-world")
            log.Infof(" >>> Hello, world!")
            return nil
    }
    ```

4.  Run `make` again. The plugin `hello-world` is ready now.

### Example 2: Generate Plugins with Input and Output Schemas

Let's add 2 plugins, a `hello-world-with-output` to generate a message, and
a `hello-world-with-input` to receive the message.

1.  Define the schema for the message. Let's create a schema file
    called `api/schemas/plugins/message.yml`:

    *NOTE:* The name of the schema definition must be the same as the file 
        name.

    ```
    definitions:
      message:
        type: object
        properties:
          words:
            type: string
    ```

2.  Edit `pkg/epplugins/plugins.yml` to define the new plugins:

    ```
    - name: hello-world-with-output
      output:
      - name: mymessage
        schema: api/schemas/plugins/message.yml

    - name: hello-world-with-input
      input:
      - name: mymessage
        schema: api/schemas/plugins/message.yml
    ```

3.  `make` the codebase. The code for the message schema and the 2 plugins will
    be generated automatically:

    ```
    pkg/api/plugins/
    ├── ...
    ├── message.go
    └── ...
    pkg/epplugins/hello-world-with-input
    ├── generated.go
    ├── generated_testutil.go
    ├── main.go
    └── main_test.go
    pkg/epplugins/hello-world-with-output
    ├── generated.go
    ├── generated_testutil.go
    ├── main.go
    └── main_test.go
    ```

4.  Edit the `PluginMain` functions of the 2 plugins:

    * `PluginMain` in `pkg/epplugins/hello-world-with-output/main.go`

        ```
        func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
                output_mymessage := output_mymessage(outp)

                // Add Plugin Code Here
                log.Infof(" >>> Plugin: hello-world-with-output")
                // Follow the data structure generated in pkg/api/plugins/message.go,
                //  which is auto-generated from the schema file message.yml.
                log.Infof(" >>> Sending Message...")
                output_mymessage.Words = "Hello World from hello-world-with-output"

                return nil
        }
        ```

    * `PluginMain` in `pkg/epplugins/hello-world-with-input/main.go`

        ```
        func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
                input_mymessage := input_mymessage(in)

                // Add Plugin Code Here
                log.Infof(" >>> Plugin: hello-world-with-input")
                // Follow the data structure generated in pkg/api/plugins/message.go,
                //  which is auto-generated from the schema file message.yml.
                log.Infof(" >>> Received Message: %s", input_mymessage.Words)

                return nil
        }
        ```

5.  Run `make` again. The plugins `hello-world-with-output` and
    `hello-world-with-input` are ready now.

### Example 3: Connect Plugins into an Existing Workflow

Let's add the plugins we created into the `init` workflow, which will be
launched when we run the `./conductor init` command.

1.  Edit `configs/workflow/common/init.yml` to add the 3 plugins in the `init` workflow as
    below, then run `make` again:

    ```
    apiVersion: conductor/v1
    kind: Workflow
    metadata:
      name: conductor-workflow
      namespace: edgeconductor
    spec:
      workflows:
      - name: init
        steps:
    {{ if eq .Kitconfig.Parameters.Customconfig.Registry.Externalurl "" }}
        - name: docker-run
          input:
          - name: containers-harbor
            schema: containers
    {{ end }}
        ## Add test plugins here
        - name: hello-world
        - name: hello-world-with-output
          output:
          - name: hello-message
            schema: mymessage
        - name: hello-world-with-input
          input:
          - name: hello-message
            schema: mymessage
    ```
2. Edit `_workspace/kit/kind.yml`

    *NOTE:* the Edge Conductor Kit config file(kit/kind.yml) is a mandatory parameter for `conductor init`. For more details of Kit config, please follow the instruction of [Custom Config](cluster-deploy-KIND.md#L38) as reference.
2.  Enter the `_workspace` folder and run `./conductor init`:

    ```
    cd ~/edge-conductor/_workspace
    ./conductor init -c kit/kind.yml
    ```

When we look at the log messages below, the 3 plugins are called
one-by-one, and the message content is successfully passed from
`hello-world-with-output` to `hello-world-with-input`.

```
...
INFO[0007] kickoff plugin: hello-world
INFO[0007] PluginConnect: plugin hello-world is connected
INFO[0007] Connected Plugin hello-world
INFO[0007] Exec Plugin hello-world
INFO[0007]  >>> Plugin: hello-world
INFO[0007]  >>> Hello, world!
INFO[0007] Complete Plugin hello-world
INFO[0007] PluginComplete: plugin hello-world, res Success
INFO[0007] kickoff plugin: hello-world-with-output
INFO[0007] PluginConnect: plugin hello-world-with-output is connected
INFO[0007] Connected Plugin hello-world-with-output
INFO[0007] Exec Plugin hello-world-with-output
INFO[0007]  >>> Plugin: hello-world-with-output
INFO[0007]  >>> Sending Message...
INFO[0007] Complete Plugin hello-world-with-output
INFO[0007] PluginComplete: plugin hello-world-with-output, res Success
INFO[0007] kickoff plugin: hello-world-with-input
INFO[0007] PluginConnect: plugin hello-world-with-input is connected
INFO[0007] Connected Plugin hello-world-with-input
INFO[0007] Exec Plugin hello-world-with-input
INFO[0007]  >>> Plugin: hello-world-with-input
INFO[0007]  >>> Received Message: Hello World from hello-world-with-output
INFO[0007] Complete Plugin hello-world-with-input
INFO[0007] PluginComplete: plugin hello-world-with-input, res Success
INFO[0007] workflow finished
```

These simple examples should give you a basic understanding of how Edge Conductor
uses plugins and workflows, and provide a foundation for more complex development. 

Copyright (c) 2022 Intel Corporation

SPDX-License-Identifier: Apache-2.0
