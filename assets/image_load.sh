#!/bin/bash

NUM_OF_TAGS=${NUM_OF_TAGS:-1}
RATE=${RATE:-16}
LOAD_REPO=${LOAD_REPO:-"quay.io/vchalla/clair-load-test"}
IMAGES=${IMAGES:-"quay.io/clair-load-test/ubuntu:xenial,\
quay.io/clair-load-test/ubuntu:focal,\
quay.io/clair-load-test/ubuntu:impish,\
quay.io/clair-load-test/ubuntu:trusty,\
quay.io/clair-load-test/ubuntu:bionic,\
quay.io/clair-load-test/alpine:3.14.0,\
quay.io/clair-load-test/busybox:uclibc,\
quay.io/clair-load-test/postgres:9.6.22,\
quay.io/clair-load-test/redis:buster,\
quay.io/clair-load-test/mysql:8.0.25,\
quay.io/clair-load-test/mongo:5.0.0-rc3,\
quay.io/clair-load-test/golang:1.18beta2,\
quay.io/clair-load-test/golang:1.17rc1,\
quay.io/clair-load-test/consul:1.9.7,\
quay.io/clair-load-test/influxdb:1.8.6,\
quay.io/clair-load-test/memcached:alpine3.14,\
quay.io/clair-load-test/centos:7,\
quay.io/clair-load-test/rabbitmq:3.9-rc,\
quay.io/clair-load-test/elasticsearch:7.13.3,\
quay.io/clair-load-test/ghost:4.9.4,\
quay.io/clair-load-test/redmine:4.2.1,\
quay.io/clair-load-test/nats:2.3.2,\
quay.io/clair-load-test/notary:signer-0.6.1-2,\
quay.io/clair-load-test/odoo:14,\
quay.io/clair-load-test/debian:bullseye,\
quay.io/clair-load-test/debian:sid,\
quay.io/clair-load-test/debian:stretch,\
quay.io/clair-load-test/kong:2.5.0,\
quay.io/clair-load-test/node:latest,\
quay.io/clair-load-test/nginx:mainline,\
quay.io/clair-load-test/couchbase:enterprise-6.6.2,\
quay.io/clair-load-test/traefik:v2.5,\
quay.io/clair-load-test/maven:openjdk,\
quay.io/clair-load-test/ruby:3.0.2,\
quay.io/clair-load-test/wordpress:php8.0,\
quay.io/clair-load-test/nextcloud:fpm,\
quay.io/clair-load-test/zookeeper:3.7.0,\
quay.io/clair-load-test/telegraf:1.19.1,\
quay.io/clair-load-test/vault:1.7.3,\
quay.io/clair-load-test/solar:8.9.0,\
quay.io/clair-load-test/joomla:php7.4,\
quay.io/clair-load-test/hadoop:latest,\
quay.io/clair-load-test/quay-rhel8:v3.6.4-2,\
quay.io/clair-load-test/debian:buster"}
# Set the Dockerfile contents
unique_id=$(cat /proc/sys/kernel/random/uuid)
for image in ${IMAGES//,/ }; do
# Extracts tag lastword
tag_prefix=$(basename "$image")
lastword=${tag_prefix##*/}
lastword=${lastword/:/_}
# Set the tag name and upload them (uses multiprocessing)
seq 1 $NUM_OF_TAGS | xargs -I {} -P $RATE bash -c '
  i="$1"
  # unique docker file to have unique manifest
  dockerfile=$(cat << EOF
FROM $4
RUN echo $5$i > /tmp/key.txt
EOF
)
  tag_name="$2:$3_tag_$i"
  # Build the Docker image using Podman
  echo "$dockerfile" | podman build \
    --tag "$tag_name" \
    --storage-opt "overlay.mount_program=/usr/bin/fuse-overlayfs" \
    --storage-driver overlay \
    -
  # Push the Docker image to a registry using Podman
  podman push "$tag_name" \
    --tls-verify=false \
    --storage-opt "overlay.mount_program=/usr/bin/fuse-overlayfs" \
    --storage-driver overlay
' _ {} "$LOAD_REPO" "$lastword" "$image" "$unique_id" &
done
# Note: Use the below command to kill this process.
# sudo pkill -f 'podman.*--tag'
# Sample execution: NUM_OF_TAGS=100000 IMAGES="quay.io/clair-load-test/mysql:8.0.25" LOAD_REPO="quay.io/vchalla/clair-load-test" RATE=20 bash image_load.sh