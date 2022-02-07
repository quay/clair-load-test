#!/usr/bin/env bash
MODE=${MODE:-$1}
MODE=${MODE:-run}
IMAGES=${IMAGES:-"quay.io/clair-load-test/ubuntu:xenial,\
quay.io/clair-load-test/alpine:3.14.0,\
quay.io/clair-load-test/busybox:uclibc,\
quay.io/clair-load-test/postgres:9.6.22,\
quay.io/clair-load-test/redis:buster,\
quay.io/clair-load-test/mysql:8.0.25,\
quay.io/clair-load-test/mongo:5.0.0-rc3,\
quay.io/clair-load-test/golang:1.18beta2,\
quay.io/clair-load-test/consul:1.9.7,\
quay.io/clair-load-test/influxdb:1.8.6,\
quay.io/clair-load-test/memcached:alpine3.14,\
quay.io/clair-load-test/centos:7,\
quay.io/clair-load-test/rabbitmq:3.9-rc,\
quay.io/clair-load-test/elasticsearch:7.13.3,\
quay.io/clair-load-test/kong:2.5.0,\
quay.io/clair-load-test/ghost:4.9.4,\
quay.io/clair-load-test/redmine:4.2.1,\
quay.io/clair-load-test/nats:2.3.2,\
quay.io/clair-load-test/notary:signer-0.6.1-2,\
quay.io/clair-load-test/odoo:14,\
quay.io/clair-load-test/debian:buster,\
quay.io/clair-load-test/ubuntu:focal"}
HOST=${HOST:-"http://localhost:6060"}
DELETE=${DELETE:0}

display_usage() {
    echo "Usage: ${0} <shell|run|help>"
    echo
    echo "The first argument needs to be one of the above modes"
    echo "if no argument is supplied the default is \"run\"."
    echo "\"help\" displays this message"
}

if [[ "${MODE}" = "help" ]]
then
    display_usage
    exit 0
fi

index() {
    time clair-load-test -D report \
        --containers ${IMAGES} \
        --host=${HOST} \
        --delete=${DELETE} \
        --psk="${PSK}"
}

case "$MODE" in
    "shell")
        echo "Entering shell mode"
        exec /bin/bash
        ;;
    "run")
        echo "Running loadtest"
         index 2>&1
        ;;
esac
