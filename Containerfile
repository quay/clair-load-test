FROM registry.fedoraproject.org/fedora-minimal:latest
RUN microdnf install rsync -y && rm -Rf /var/cache/yum
COPY clair-load-test /bin/clair-load-test
LABEL io.k8s.display-name="clair-load-test"
ENTRYPOINT ["/bin/clair-load-test -D report"]