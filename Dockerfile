FROM quay.io/projectquay/golang:1.15 AS build
WORKDIR /build/
ADD . /build/
RUN go build

FROM quay.io/projectquay/clair:4.3.6
WORKDIR /run
COPY entrypoint.sh entrypoint.sh
COPY --from=build /build/clair-load-test /bin/clair-load-test
ENTRYPOINT ["/run/entrypoint.sh"]
