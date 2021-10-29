REGISTRY := docker.io
REPO_NAME ?= nmap_prometheus
PACKAGE_NAME := github.com/beaujr/$(REPO_NAME)
APP_NAME := beaujr/$(REPO_NAME)
IMAGE_TAG ?= 0.1
GOPATH ?= $HOME/go
APP_TYPE := client
BUILD_TAG := build-$(APP_TYPE)
BINPATH := ../bin
NAMESPACE := default
PORT := 1234
PLATFORM := linux/amd64,linux/arm/v7,linux/arm/v6,linux/arm64

# Path to dockerfiles directory
DOCKERFILES := build

GIT_COMMIT := $(shell git rev-parse --short HEAD)
GOLDFLAGS := -ldflags "-X $(PACKAGE_NAME)/pkg/util.AppGitCommit=${GIT_COMMIT} -X $(PACKAGE_NAME)/pkg/util.AppVersion=${IMAGE_TAG}"

.PHONY: verify build docker_build push generate generate_verify \
	go_fcm_server go_test go_fmt e2e_test go_verify   \
	docker_push

# Alias targets
###############

build: go_mod go_test nmap_prometheus # docker_build
verify: generate_verify go_verify
#push: build docker_push

# Go targets
#################
go_verify: go_fmt go_test

go_mod:
	go mod tidy
	go mod vendor

nmap_prometheus:
	cd $(APP_TYPE) && \
	CGO_ENABLED=0 go build  \
		-a -tags netgo \
		-o $(BINPATH)/${APP_NAME}-$(APP_TYPE) \
		./
#	sftp -o "IdentityFile=~/.ssh/raspberry" -oPort=22 nuc@192.168.1.112:/home/nuc/repo <<< $$'put bin/beaujr/nmap_prometheus-client'

go_test:
ifeq ($(GOARCH),amd64)
	CGO_ENABLED=0 go test -v \
		-cover \
		-coverprofile=coverage.out \
		$$(go list ./... | \
			grep -v '/vendor/' | \
			grep -v '/pkg/client' \
		)
endif

coverage: go_test
	go tool cover -html=coverage.out

go_fmt:
	@set -e; \
	GO_FMT=$$(git ls-files *.go | grep -v 'vendor/' | xargs gofmt -d); \
	if [ -n "$${GO_FMT}" ] ; then \
		echo "Please run go fmt"; \
		echo "$$GO_FMT"; \
		exit 1; \
	fi

docker-build-and-push: PUSH=true
docker-build-and-push: TYPE=registry
docker-build-and-push: BUILD_TAG=$(IMAGE_TAG)-$(APP_TYPE)-$(GIT_COMMIT)
docker-build-and-push: | docker-login docker_build

docker_build: DOCKERFILE=Dockerfile
docker_build: PUSH=true
docker_build: TYPE=image
docker_build:
	docker buildx build \
		--build-arg VCS_REF=$(GIT_COMMIT) \
		--build-arg APP_TYPE=$(APP_TYPE) \
		--build-arg APP_NAME=$(REPO_NAME) \
		--tag $(REGISTRY)/$(APP_NAME):$(BUILD_TAG) \
		--platform $(PLATFORM) \
		--output "type=$(TYPE),push=$(PUSH)" \
		--file $(DOCKERFILES)/$(DOCKERFILE) \
		./

docker_run:
	@docker run -p $(PORT):$(PORT) -v $(shell pwd)/config:/config $(REGISTRY)/$(APP_NAME):$(BUILD_TAG) -port=$(PORT)

docker_push: docker-login
	set -e; \
	docker tag $(REGISTRY)/$(APP_NAME):$(BUILD_TAG) $(APP_NAME):$(IMAGE_TAG)-$(APP_TYPE)-$(GOARCH)-$(GIT_COMMIT) ; \
	docker push $(APP_NAME):$(IMAGE_TAG)-$(APP_TYPE)-$(GOARCH)-$(GIT_COMMIT);
ifeq ($(GITHUB_HEAD_REF),master)
	docker tag $(APP_NAME):$(IMAGE_TAG)-$(GOARCH)-$(GIT_COMMIT) $(APP_NAME):latest_$(GOARCH)
	docker push $(APP_NAME):latest_$(GOARCH)
endif

check-docker-credentials:
ifndef DOCKER_USER
	$(error DOCKER_USER is undefined)
else
  ifndef DOCKER_PASS
	$(error DOCKER_PASS is undefined)
  endif
endif

grpcUrl-device:
	@grpcurl -plaintext -d '{"addresses": [{"ip":"192.168.1.220", "mac":"BE:6D:C8:C9:2B:0B"}]}' -H "Host: grpc.beau.cf" -H "Home: aus" -H "Client: mbp" 192.168.1.112:50051 proto.HomeDetector.Addresses

grpcUrl-ble:
	@grpcurl -v -plaintext -d '{ "key":"0c:f3:ee:04:36:c0"}' -use-reflection=false -proto=proto/DeviceDetector.proto -H="Content-Type: application/grpc" -H "Home: aus" -H "Client: mbp"  192.168.1.112:50051 proto.HomeDetector.Ack

#grpcurl -v -plaintext -d '{ "id": "1", "executeat": 2, "owner": "3", "command": "turn salt on", "executed": false }' -use-reflection=false -proto=proto/DeviceDetector.proto -H="Content-Type: application/grpc" -H "Home: aus" -H "Client: mbp"  192.168.1.112:50051 proto.HomeDetector.CreateTimedCommand
docker-login: check-docker-credentials
	@docker login -u $(DOCKER_USER) -p $(DOCKER_PASS) $(REGISTRY)

score: PR_ID=$(shell echo $(GITHUB_REF) | tr -dc '0-9')
score:
	curl -X GET \
	https://gogitops.beau.cf/submit/$(GITHUB_REPOSITORY)/pull/$(PR_ID) \
	-H 'user: $(GITHUB_USER)' \
	-H 'token: $(GITHUB_TOKEN)'

proto:
	protoc -I $GOPATH/pkg/mod/github.com/srikrsna/protoc-gen-gotag@v0.5.0/ --proto_path=proto --go_out=proto/ --go-grpc_out=proto/ proto/DeviceDetector.proto