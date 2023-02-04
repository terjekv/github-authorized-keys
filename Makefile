# Suported platforms for compilation
RELEASE_ARCH += darwin/386
RELEASE_ARCH += darwin/amd64
RELEASE_ARCH += linux/386
RELEASE_ARCH += linux/amd64
RELEASE_ARCH += linux/arm
RELEASE_ARCH += linux/arm64
RELEASE_ARCH += freebsd/386
RELEASE_ARCH += freebsd/amd64
RELEASE_ARCH += freebsd/arm
RELEASE_ARCH += netbsd/386
RELEASE_ARCH += netbsd/amd64
RELEASE_ARCH += netbsd/arm
RELEASE_ARCH += openbsd/386
RELEASE_ARCH += openbsd/amd64

APP := github-authorized-keys

BUILD_FOR := arm64 amd64

COPYRIGHT_SOFTWARE := Github Authorized Keys
COPYRIGHT_SOFTWARE_DESCRIPTION := Use GitHub teams to manage system user accounts and authorized_keys

export DOCKER_IMAGE_NAME = terjekv/$(APP)

# include $(shell curl -so .build-harness "https://raw.githubusercontent.com/cloudposse/build-harness/master/templates/Makefile.build-harness"; echo .build-harness)

## Execute local build
build:
	for arch in $(BUILD_FOR); do GOOS=linux GOARCH=$$arch go build -o release/github-authorized-keys.$${arch}; done
##	$(SELF) go:build 

## Execute local deps
deps:
	go get .

clean:
	rm -rf release

go:
	$(SELF) build
	
## Docker buildx
docker:
	docker buildx build --platform=linux/arm64,linux/amd64 .

## Execute all targets
all:
#	 $(SELF) go:deps-dev
#	 $(SELF) go:deps-build
	 $(SELF) deps 
#	 $(SELF) go:lint
#	 $(SELF) go:test 
#	 $(SELF) go:build-all
	$(SELF) build

## Bring up docker compose environment
compose-up:
	docker-compose -f docker-compose-test.yaml up -d
