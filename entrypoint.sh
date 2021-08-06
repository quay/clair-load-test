#!/usr/bin/env bash
LOADTESTENTRY=${LOADTESTENTRY:-$1}
LOADTESTENTRY=${LOADTESTENTRY:-short}
IMAGES=${IMAGES:-"quay.io/crozzy/ubuntu:xenial,quay.io/crozzy/alpine:3.14.0,quay.io/crozzy/busybox:uclibc,quay.io/crozzy/postgres:9.6.22,quay.io/crozzy/redis:buster,quay.io/crozzy/node:latest,quay.io/crozzy/mysql:8.0.25,quay.io/crozzy/mongo:5.0.0-rc3,quay.io/crozzy/nginx:mainline,quay.io/crozzy/couchbase:enterprise-6.6.2,quay.io/crozzy/traefik:v2.5,quay.io/crozzy/golang:1.17rc1,quay.io/crozzy/consul:1.9.7,quay.io/crozzy/influxdb:1.8.6,quay.io/crozzy/memcached:alpine3.14,quay.io/crozzy/centos:7,quay.io/crozzy/maven:openjdk,quay.io/crozzy/ruby:3.0.2,quay.io/crozzy/wordpress:php8.0,quay.io/crozzy/rabbitmq:3.9-rc,quay.io/crozzy/elasticsearch:7.13.3,quay.io/crozzy/nextcloud:fpm,quay.io/crozzy/zookeeper:3.7.0,quay.io/crozzy/telegraf:1.19.1,quay.io/crozzy/vault:1.7.3,quay.io/crozzy/kong:2.5.0,quay.io/crozzy/ghost:4.9.4,quay.io/crozzy/solr:8.9.0,quay.io/crozzy/joomla:php7.4,quay.io/crozzy/redmine:4.2.1,quay.io/crozzy/nats:2.3.2,quay.io/crozzy/notary:signer-0.6.1-2,quay.io/crozzy/odoo:14"}
HOST=${HOST:-"http://localhost:6060"}
CONCURRENCY=${CONCURRENCY:-10}

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
    time clair-load-test -D -c ${CLAIR_CONF} report --containers ${IMAGES} --concurrency=${CONCURRENCY} --host="${HOST}"
}

flushdb() {
    time clair-load-test -D -c ${CLAIR_CONF} flushdb -y
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
    "flushdb")
        echo "Flushing DB"
        flushdb
        ;;
    "prolonged")
        echo "Running prolonged load test"
        while :
        do 
            index 2>&1
            flushdb 2>&1
        done
        ;;
esac