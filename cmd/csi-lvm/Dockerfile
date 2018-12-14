FROM golang AS builder
RUN  go get github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/steigr/csi-lvm
COPY Gopkg.toml Gopkg.toml
COPY Gopkg.lock Gopkg.lock
COPY vendor vendor
COPY Makefile Makefile
COPY cmd cmd
COPY pkg pkg
COPY .git .git
RUN  make csi-lvm

FROM alpine
LABEL maintainers="Mathias Kaufmann"
LABEL description="LVM Driver"

COPY --from=builder /go/src/github.com/steigr/csi-lvm/_output/lvmplugin /bin/lvmplugin
ENTRYPOINT ["lvmplugin"]