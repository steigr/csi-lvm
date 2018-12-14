REGISTRY_NAME=quay.io/steigr
IMAGE_NAME=lvmplugin
IMAGE_VERSION=canary
IMAGE_TAG=$(REGISTRY_NAME)/$(IMAGE_NAME):$(IMAGE_VERSION)
REV=$(shell git describe --long --tags --dirty)

csi-lvm:
	if [ ! -d ./vendor ]; then dep ensure -vendor-only; fi
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-X github.com/steigr/csi-lvm/pkg/lvm.vendorVersion=$(REV) -extldflags "-static"' -o _output/lvmplugin ./cmd/csi-lvm
