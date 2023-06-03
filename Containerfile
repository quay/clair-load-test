FROM ubuntu

# setup installation directory and disable prompts during installation
WORKDIR /tmp
ARG DEBIAN_FRONTEND=noninteractive

# Install necessary libraries & cleanup for subsequent commands
RUN apt-get update && apt-get install -y wget dumb-init \
 && apt-get clean autoclean \
 && apt-get autoremove --yes \
 && rm -rf /var/lib/{apt,dpkg,cache,log}/

# Install clairctl to get clair manifests
RUN wget https://github.com/quay/clair/releases/download/v4.6.0/clairctl-linux-amd64 \
 && chmod 777 clairctl-linux-amd64 \
 && mv clairctl-linux-amd64 /usr/local/bin/clairctl \
 && rm -rf clairctl-linux-amd64

# Copy binary and start the command
COPY clair-load-test /bin/clair-load-test
LABEL io.k8s.display-name="clair-load-test"
ENTRYPOINT ["/usr/bin/dumb-init", "--", "sh", "-c", "ulimit -n 65536 && ulimit -p 65536 && clair-load-test -D report"]