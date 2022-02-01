# Clair Load Testing

This project provides a simple CLI for making requests to Clair. Although it doesn't boast the same HTTP control as a load testing tool such as [wrk](https://github.com/wg/wrk), it does offer a way to construct API calls to Clair that all container layers to be fetched without the need for Quay. This tool is build on top of [clairctl](https://github.com/quay/clair/blob/cbdc9caab450489377ab1d6bb19429d54df639cc/Documentation/reference/clairctl.md) and requires it in your path.

> **NOTE**: `clair-load-test` is **NOT** for use on production instances of Clair, it does some not so nice things to the database.

## Prerequisites

* A running instance of Clair (to test).
* `clairctl` in your path.

## Usage
```
NAME:
   clair-load-test - Stress your Clair

USAGE:
   clair-load-test [global options] command [command options] [arguments...]

VERSION:
   0.0.1

DESCRIPTION:
   A command-line tool for stress testing clair v4.

COMMANDS:
   report       clair-load-test report
   flushdb      clair-load-test flushdb
   createtoken  createtoken --key sdfvevefr==
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -D                        print debugging logs (default: false)
   -q                        quieter log output (default: false)
   --config value, -c value  clair configuration file (default: "config.yaml") [$CLAIR_CONF]
   --help, -h                show help (default: false)
   --version, -v             print the version (default: false)

```

> **NOTE**: The config schema is the same as for a running instance of Clair, this allows the tool to authenticate to execute both HTTP requests and Database operations.

## Installation

```
make build
```

## Examples

### Index and retrieve the Vulnerability Report for some images from a local Clair instance, 5 at a time:
```sh
clair-load-test -D report --containers ubuntu:xenial,alpine:3.14.0,busybox:uclibc,postgres:9.6.22,redis:buster,python:slim,node:latest,mysql:8.0.25,mongo:5.0.0-rc3,nginx:mainline --concurrency=5 --host="http://localhost:6060"
```

### Flush Database tables to ensure reindexing on resubmission:
```
clair-load-test -D -c ./loadtest-local.yaml flushdb
```

## Containerized Running

In the interests of making the tool portable and dependency free (well almost). It is possible to run in a container.

### Build and run the container

```sh
podman build . -t clair-load-test
podman run -v config:/config -e CONCURRENCY=10 -e HOST=http://clair-indexer-perftestx-clair.apps.quaydev-rosa-1.czz9.p1.openshiftapps.com -e CLAIR_CONF=/config/loadtest-dist.yaml -e LOADTESTENTRY=short -it clair-load-test
```

