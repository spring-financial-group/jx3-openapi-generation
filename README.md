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

To run the package locally you will need to do few prep steps before you can run the generation.

Open the service you wish to generate the packages for and ensure you have an OpenAPI specification file.

Then set the required environment variables in your shell (or use an `.envrc` file with [direnv](https://direnv.net/) in the target repo folder).

The following environment variables are required:

```
VERSION="1.0.0"
REPO_OWNER="spring-financial-group"
REPO_NAME="mqube-something-service"
SwaggerServiceName="SomethingService"
SpecPath="./docs/openapi.json"
GIT_USER="your-git-username"
GIT_TOKEN="your-git-token"
```

Then copy the `configs` directory from this repository to the root of the service repository.

Next, open the `pkg/openapitools/config.go` and change the `ConfigsDir` to `"./configs"`.
Since you want to run it locally you most likely want to view the generated packages, to do that you will need to comment out a line in `pkg/cmd/generate/generate_packages.go` that removes the temporary directory after generation look for `defer o.FileIO.DeferRemove(tmpDir)` in the `Run()` function.

Each language generator has its own push logic, which will use your credentials to create a commit and push the generated package to the relevant repository. You want to ensure you have that code commented out before running the package generation locally, otherwise you will end up pushing - possibly - incompatible packages to the repositories.

FINALLY. You are now ready to build the CLI. Run the following command to build the CLI in this repository folder:

```bash
make build
```

This will create the `jx3-openapi-generation` binary in the `build` directory.

Copy the `jx3-openapi-generation` to the root folder of the target repository.

Then run the following command to generate the packages:

```bash
./jx3-openapi-generation generate pkg python
```

You may need to run `chmod +x ./jx3-openapi-generation` to make the binary executable first.
