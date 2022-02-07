# Clair Load Testing

This project provides a simple CLI for making requests to Clair. Although it doesn't boast the same HTTP control as a load testing tool such as [wrk](https://github.com/wg/wrk), it does offer a way to construct API calls to Clair that all container layers to be fetched without the need for Quay. This tool uses [clairctl](https://github.com/quay/clair/blob/cbdc9caab450489377ab1d6bb19429d54df639cc/Documentation/reference/clairctl.md) to create manifest definitions and requires it in your path.

> **NOTE**: `clair-load-test` is **NOT** for use on production instances of Clair.

## Prerequisites

* A running instance of Clair (to test).
* `clairctl` in your path.

## Usage
```
NAME:
   clair-load-test - A command-line tool for stress testing clair v4.

USAGE:
   clair-load-test [global options] command [command options] [arguments...]

VERSION:
   0.0.1

DESCRIPTION:
   A command-line tool for stress testing clair v4.

COMMANDS:
   report       clair-load-test report
   createtoken  createtoken --key sdfvevefr==
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   -D             print debugging logs (default: false)
   -q             quieter log output (default: false)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)

```

### Report
```
NAME:
   clair-load-test report - clair-load-test report

USAGE:
   clair-load-test report [command options] [arguments...]

DESCRIPTION:
   request reports for named containers

OPTIONS:
   --host value         --host localhost:6060/ (default: "http://localhost:6060/") [$CLAIR_API]
   --containers value   --containers ubuntu:latest,mysql:latest [$CONTAINERS]
   --psk value          --psk secretkey [$PSK]
   --delete             --delete (default: false) [$DELETE]
   --timeout value      --timeout 1m (default: 1m0s) [$TIMEOUT]
   --rate value         --rate 1 (default: 1) [$RATE]
   --help, -h           show help (default: false)
```

## Installation

```
make build
```

## Examples

### Index and retrieve the Vulnerability Report for some images from a local Clair instance, at a rate of 1 per second and delete the Index Report after:
```sh
clair-load-test -D report --containers ubuntu:xenial,alpine:3.14.0,busybox:uclibc,postgres:9.6.22,redis:buster,python:slim,node:latest,mysql:8.0.25,mongo:5.0.0-rc3,nginx:mainline --rate=1 --host="http://localhost:6060" --psk=secret --timeout=2m --delete=1
```

## Containerized Running

In the interests of making the tool portable and dependency free (well almost). It is possible to run in a container.

### Build and run the container

```sh
podman build . -t clair-load-test
podman run -e HOST=http://clair-indexer-perftestx-clair.apps.quaydev-rosa-1.czz9.p1.openshiftapps.com -e TIMEOUT=1m -e DELETE=1 -e PSK=secret -e RATE=1 -it clair-load-test
```

