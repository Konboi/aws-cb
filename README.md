# aws-cb

AWS CodeBuild utility tool

## Install

```sh
# install from source
go install github.com/Konboi/aws-cb/cmd/cb@latest

# or build in a local checkout
go build -o cb ./cmd/cb
```

## Usage

```
cb - AWS CodeBuild utility tool

Usage:
  cb list                    # list CodeBuild projects
  cb list <PROJECT>          # list recent builds for a project
  cb less <BUILD_ID>         # display build log (stub)
  cb rerun <BUILD_ID>        # rerun build (stub)

Flags: (command-specific where applicable)
  -h, --help                 Show help
```
