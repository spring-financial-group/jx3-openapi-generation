## jx3-openapi-generation

This repository contains code and configuration for generating client packages using the OpenAPI Generator tool.
The tool is wrapped in a Go CLI which is used for setting up the configuration for each language and pushing the generated
packages to the relevant repositories.

The following languages currently are supported:

| Language   | Argument     | Notes                                    |
| ---------- | ------------ | ---------------------------------------- |
| C#         | `csharp`     |                                          |
| Java       | `java`       |                                          |
| Angular    | `angular`    |                                          |
| Typescript | `typescript` |                                          |
| Python     | `python`     | **Not available for preview packages\*** |
| Golang     | `go`         | **Not available for preview packages\*** |
| Rust       | `rust`       |                                          |

\* Due to the way in which Go and Python packages are stored preview packages are not current supported so avoid
using these languages in any preview pipelines

## Usage

The CLI is configured through environment variables. The following environment variables are required:

| Variable Name        | Description                                                                                   |
| -------------------- | --------------------------------------------------------------------------------------------- |
| `SwaggerServiceName` | The name of the service to be used to generate .                                              |
| `SpecPath`           | The path to the OpenAPI spec file. This is relative to the root of the repository.            |
| `VERSION`            | The semvar version of the service. Used to keep the package version in step with the service. |
| `REPO_OWNER`         | The owner of the service repository.                                                          |
| `REPO_NAME`          | The name of the service repository.                                                           |
| `GIT_TOKEN`          | Authorisation token used for pushing Python packages to a repository.                         |
| `GIT_USER`           | The user to use for authenticating with GitHub                                                |

Then to generate a package for a service, run the following command:

```bash
jx3-openapi-generation generate <languages>
```

where `<languages>` is a space-separated list of languages to generate packages for.

### Jenkins X

To call the CLI from a Jenkins X pipeline, add the following as the final step in the `release.yaml` or `pullrequest.yaml`
pipelines:

```yaml
- image: uses:spring-financial-group/jx3-openapi-generation/pipeline/generate-packages.yaml@master
  name: ""
  resources: {}
```

The languages to generate packages for are configured by setting the environment variable `OutputLanguages` in the
environment variables of the pipeline. The `SwaggerServiceName` & `SpecPath` variables are also required.

Note that the other environment variables are set by default in JX pipelines so not required in your definition.

```yaml
env:
  - name: SwaggerServiceName
    value: PetStoreService
  - name: SpecPath
    value: ./docs/swagger.json
  - name: OutputLanguages
    value: csharp angular java go typescript
```

#### Full Example

```yaml
apiVersion: tekton.dev/v1beta1
kind: PipelineRun
metadata:
  name: pullrequest
spec:
  pipelineSpec:
    tasks:
    - name: from-build-pack
      resources: {}
      taskSpec:
        metadata: {}
        stepTemplate:
          image: uses:jenkins-x/jx3-pipeline-catalog/tasks/go/pullrequest.yaml@versionStream
          name: ""
          resources:
            limits: {}
          workingDir: /workspace/source
          env:
          - name: SwaggerServiceName
            value: PetStoreService
          - name: SpecPath
            value: ./docs/swagger.json
          - name: OutputLanguages
            value: csharp angular java
        steps:
        - image: uses:jenkins-x/jx3-pipeline-catalog/tasks/git-clone/git-clone-pr.yaml@versionStream
          name: ""
          resources: {}
        - name: jx-variables
          resources: {}
        - name: build-make-build
          resources: {}
        - name: check-registry
          resources: {}
        - name: build-scan-push
          resources: {}
        - name: promote-jx-preview
          resources: {}
        - image: uses:spring-financial-group/jx3-openapi-generation/pipeline/generate-packages.yaml@master
          name: ""
          resources: {}
  podTemplate: {}
  serviceAccountName: tekton-bot
  timeout: 1h0m0s
status: {}
```

## Running Locally

To generate packages locally you only need an OpenAPI specification file for the service you wish to generate packages for — no environment variables, credential setup, or repository cloning is required.

First, build the CLI from this repository:

```bash
make build
```

This creates the `jx3-openapi-generation` binary in the `build` directory.

Then generate packages for the languages you want, skipping the push step and writing the output to a local directory:

```bash
./build/jx3-openapi-generation generate packages [generators] --skip-push --spec-path=path/to/openapi.json --out=./tmp-pkg
```

where `[generators]` is a space-separated list of languages to generate packages for (e.g. `csharp java go`). The generated packages are written into the relative `./tmp-pkg` directory, grouped by language.

With `--skip-push` set, only `--spec-path` is required — no Git credentials or repository environment variables are needed, and nothing is pushed to any remote. The configs are bundled with the CLI, so there is no need to copy a `configs` directory or edit any source files.
