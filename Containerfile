FROM ubuntu

LABEL maintainer="vchalla@redhat.com"

WORKDIR /tmp
ARG DEBIAN_FRONTEND=noninteractive

# ALPINE
RUN apt-get update
RUN apt-get upgrade  -y
RUN apt-get install -y podman
RUN apt-get install -y wget python3.6 git dumb-init

# Snafu Build Dependencies
# TODO: A lot of the *-dev dependencies were added while attempting to build
#       snafu's requirements (mainly numpy, scipy, pandas). Now that these are
#       installed using Alpine's package manager, they may not all be necessary
#       anymore.
RUN apt-get install -y python3-numpy python3-scipy python3-pandas gcc python3-dev \
                       postgresql-client libffi-dev libxml2 \
                       libxml2-dev libxslt-dev libjpeg-dev \
                       zlib1g-dev musl-dev

# RUN apt-get install -y libressl-dev
# Install required third-party packages
RUN wget https://github.com/tsenart/vegeta/releases/download/v12.8.3/vegeta-12.8.3-linux-amd64.tar.gz
RUN tar -xzf vegeta-12.8.3-linux-amd64.tar.gz
RUN mv vegeta /usr/local/bin/vegeta

RUN wget https://github.com/quay/clair/releases/download/v4.6.0/clairctl-linux-amd64
RUN chmod 777 clairctl-linux-amd64
RUN mv clairctl-linux-amd64 /usr/local/bin/clairctl


RUN apt-get update && apt-get install -y python3.6 python3-distutils python3-pip python3-apt
RUN apt-get update && apt-get install -y redis-server
RUN apt-get update && apt-get install -y jq curl
RUN ln -s /usr/bin/python3 /usr/bin/python
RUN mkdir -p /opt/snafu/
RUN git clone https://github.com/cloud-bulldozer/benchmark-wrapper
RUN cd benchmark-wrapper && \
    cp -R . /opt/snafu/
RUN pip3 install --upgrade pip
RUN pip3 install -e /opt/snafu/

# Install Python Dependencies
# RUN python3 -m ensurepip
RUN python3 -m pip install --upgrade pip
RUN apt-get clean autoclean
RUN apt-get autoremove --yes
RUN rm -rf /var/lib/{apt,dpkg,cache,log}/
COPY clair-load-test /bin/clair-load-test
LABEL io.k8s.display-name="clair-load-test"
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["clair-load-test", "-D", "report"]