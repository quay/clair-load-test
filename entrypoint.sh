#!/usr/bin/env bash
LOADTESTENTRY=${LOADTESTENTRY:-$1}
LOADTESTENTRY=${LOADTESTENTRY:-short}
IMAGES=${IMAGES:-"quay.io/clair-load-test/ubuntu:xenial,quay.io/clair-load-test/alpine:3.14.0,quay.io/clair-load-test/busybox:uclibc,quay.io/clair-load-test/postgres:9.6.22,quay.io/clair-load-test/redis:buster,quay.io/clair-load-test/node:latest,quay.io/clair-load-test/mysql:8.0.25,quay.io/clair-load-test/mongo:5.0.0-rc3,quay.io/clair-load-test/nginx:mainline,quay.io/clair-load-test/couchbase:enterprise-6.6.2,quay.io/clair-load-test/traefik:v2.5,quay.io/clair-load-test/golang:1.17rc1,quay.io/clair-load-test/consul:1.9.7,quay.io/clair-load-test/influxdb:1.8.6,quay.io/clair-load-test/memcached:alpine3.14,quay.io/clair-load-test/centos:7,quay.io/clair-load-test/maven:openjdk,quay.io/clair-load-test/ruby:3.0.2,quay.io/clair-load-test/wordpress:php8.0,quay.io/clair-load-test/rabbitmq:3.9-rc,quay.io/clair-load-test/elasticsearch:7.13.3,quay.io/clair-load-test/nextcloud:fpm,quay.io/clair-load-test/zookeeper:3.7.0,quay.io/clair-load-test/telegraf:1.19.1,quay.io/clair-load-test/vault:1.7.3,quay.io/clair-load-test/kong:2.5.0,quay.io/clair-load-test/ghost:4.9.4,quay.io/clair-load-test/solr:8.9.0,quay.io/clair-load-test/joomla:php7.4,quay.io/clair-load-test/redmine:4.2.1,quay.io/clair-load-test/nats:2.3.2,quay.io/clair-load-test/notary:signer-0.6.1-2,quay.io/clair-load-test/odoo:14"}
HOST=${HOST:-"http://localhost:6060"}
CONCURRENCY=${CONCURRENCY:-10}
DELETE=${DELETE:0}

display_usage() {
    echo "Usage: ${0} <shell|short|prolonged|flushdb|help>"
    echo
    echo "The first argument needs to be one of the above modes"
    echo "if no argument is supplied the default is \"short\"."
    echo "\"help\" displays this message"
}

if [[ "${LOADTESTENTRY}" = "help" ]]
then
    display_usage
    exit 0
fi

index() {
    time ./clair-load-test -D report \
        --containers ${IMAGES} \
        --concurrency=${CONCURRENCY} \
        --host="${HOST}" \
        --delete=${DELETE} \
        --psk=${PSK}
}

case "$LOADTESTENTRY" in
    "shell")
        echo "Entering shell mode"
        exec /bin/bash
        ;;
    "short")
        echo "Running short loadtest"
         index 2>&1
        ;;
    "prolonged")
        echo "Running prolonged load test"
        while :
        do 
            index 2>&1
        done
        ;;
esac
