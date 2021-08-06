FROM quay.io/projectquay/golang:1.15 AS build
WORKDIR /build/
ADD . /build/
RUN go build

FROM quay.io/projectquay/clair:4.2.0-rc.2
WORKDIR /run
COPY entrypoint.sh entrypoint.sh
COPY --from=build /build/clair-load-test /bin/clair-load-test
COPY /config /config
ENTRYPOINT ["/usr/local/bin/dumb-init", "--", "/run/entrypoint.sh"]