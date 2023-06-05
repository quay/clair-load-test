#!/bin/bash

START=${START:-1}
END=${END:-10000}
RATE=${RATE:-16}
LAYERS=${LAYERS:-5}
LAYER_SUFFIX=$LAYERS
LAYERS=$((18 + ${LAYER_SUFFIX}))
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
seq $START $END | xargs -I {} -P $RATE bash -c '
  i="$1"
  # unique docker file to have unique manifest
  dockerfile=$(cat <<EOF | head -n $6
FROM registry.access.redhat.com/ubi8:latest as tarbuilder$5$i
RUN echo "This is a sample file." > my_file$5$i.txt && tar -cf my_file$5$i.tar my_file$5$i.txt && rm -f my_file$5$i.txt
FROM registry.access.redhat.com/ubi8/go-toolset:latest as gobuilder$5$i
WORKDIR /app$5$i
RUN echo -e "package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, Docker!\")\n}" > main.go
RUN go mod init my_app$5$i
RUN go build -o my_app$5$i
FROM registry.access.redhat.com/ubi8/openjdk-8:1.3 as javabuilder$5$i
WORKDIR /app$5$i
RUN echo "public class HelloWorld { public static void main(String[] args) { System.out.println(\"Hello, Docker!\"); } }" > HelloWorld.java
RUN javac HelloWorld.java && jar cfe my_app$5$i.jar HelloWorld HelloWorld.class
FROM registry.access.redhat.com/ubi8:latest AS shellbuilder$5$i
WORKDIR /app$5$i
RUN echo "#!/bin/sh" > myscript$5$i.sh && echo "echo \"Hello, world!\"" >> myscript$5$i.sh
FROM $4
WORKDIR /app$5$i
RUN apt-get update && apt-get install -y --no-install-recommends tar && rm -rf /var/lib/apt/lists/*
COPY --from=tarbuilder$5$i my_file$5$i.tar /app$5$i/
COPY --from=tarbuilder$5$i my_file$5$i.tar /build$5$i/app$5$5$i$i/
RUN tar -xf /app$5$i/my_file$5$i.tar --overwrite
RUN tar -xf /build$5$i/app$5$5$i$i/my_file$5$i.tar --overwrite
COPY --from=gobuilder$5$i /app$5$i/my_app$5$i /app$5$i/
COPY --from=gobuilder$5$i /app$5$i/my_app$5$i /build$5$i/app$5$5$i$i/
COPY --from=javabuilder$5$i /app$5$i/my_app$5$i.jar /app$5$i/
COPY --from=javabuilder$5$i /app$5$i/my_app$5$i.jar /build$5$i/app$5$5$i$i/
COPY --from=shellbuilder$5$i /app$5$i/myscript$5$i.sh /app$5$i/
COPY --from=shellbuilder$5$i /app$5$i/myscript$5$i.sh /build$5$i/app$5$5$i$i/
RUN echo $5$i > /app$5$i/key$5$i.txt
RUN echo $5$i > /build$5$i/app$5$5$i$i/key$5$i.txt
RUN chmod +x /app$5$i/my_file$5$i.tar
RUN chmod +x /build$5$i/app$5$5$i$i/my_file$5$i.tar
RUN chmod +x /app$5$i/my_app$5$i 
RUN chmod +x /build$5$i/app$5$5$i$i/my_app$5$i
RUN chmod +x /app$5$i/my_app$5$i.jar
RUN chmod +x /build$5$i/app$5$5$i$i/my_app$5$i.jar
RUN chmod +x /app$5$i/myscript$5$i.sh 
RUN chmod +x /build$5$i/app$5$5$i$i/myscript$5$i.sh
RUN chmod +x /app$5$i/key$5$i.txt
RUN chmod +x /build$5$i/app$5$5$i$i/key$5$i.txt
RUN rm -f /app$5$i/my_file$5$i.tar
RUN rm -f /build$5$i/app$5$5$i$i/my_file$5$i.tar
ENV VAR1 /app$5$i/my_file$5$i.tar
ENV VAR6 /build$5$i/app$5$5$i$i/my_file$5$i.tar
RUN rm -f /app$5$i/my_app$5$i
RUN rm -f /build$5$i/app$5$5$i$i/my_app$5$i
ENV VAR2 /app$5$i/my_app$5$i
ENV VAR7 /build$5$i/app$5$5$i$i/my_app$5$i
RUN rm -f /app$5$i/my_app$5$i.jar
RUN rm -f /build$5$i/app$5$5$i$i/my_app$5$i.jar
ENV VAR3 /app$5$i/my_app$5$i.jar
ENV VAR8 /build$5$i/app$5$5$i$i/my_app$5$i.jar
RUN rm -f /app$5$i/myscript$5$i.sh
RUN rm -f /build$5$i/app$5$5$i$i/myscript$5$i.sh
ENV VAR4 /app$5$i/myscript$5$i.sh
ENV VAR9 /build$5$i/app$5$5$i$i/myscript$5$i.sh
RUN rm -f /app$5$i/key$5$i.txt
RUN rm -f /build$5$i/app$5$5$i$i/key$5$i.txt
ENV VAR5 /app$5$i/key$5$i.txt
ENV VAR10 /build$5$i/app$5$5$i$i/key$5$i.txt
EOF
)
  tag_name="$2:$3_layers_$7_tag_$i"
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
  # Delete the Docker image using Podman
  podman rmi "$tag_name"
' _ {} "$LOAD_REPO" "$lastword" "$image" "$unique_id" "$LAYERS" "$LAYER_SUFFIX" &
done
# Note: Use the below command to kill this process.
# sudo pkill -f 'podman.*--tag'
# Sample execution: START=1 END=10000 LAYERS=5 IMAGES="quay.io/clair-load-test/mysql:8.0.25" LOAD_REPO="quay.io/vchalla/clair-load-test" RATE=20 bash image_load.sh