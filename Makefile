SHELL=/bin/bash -o pipefail

COMMIT_NUM := $(shell git rev-parse HEAD 2> /dev/null || true)
GIT_COMMIT := $(if $(shell git status --porcelain --untracked-files=no),${COMMIT_NUM}-dirty,${COMMIT_NUM})
GIT_BRANCH := $(shell echo $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null) | sed -e "s/[^[:alnum:]]/-/g")

IMAGE_BUILD_NAME ?= leodido/runbpf
IMAGE_BUILD_FLAG ?= --no-cache

IMAGE_BUILD_BRANCH := $(IMAGE_BUILD_NAME):$(GIT_BRANCH)
IMAGE_BUILD_COMMIT := $(IMAGE_BUILD_NAME):$(GIT_COMMIT)
IMAGE_BUILD_LATEST := $(IMAGE_BUILD_NAME):latest


LDFLAGS := -ldflags "" # -ldflags "-extldflags '-static'"

RUNNER ?= output/runner

$(RUNNER):
	go build ${LDFLAGS} -o $@ ./cmd/runner

$(IMAGE_BUILD_NAME):
	docker build \
		$(IMAGE_BUILD_FLAG) \
		-t $(IMAGE_BUILD_BRANCH) \
		-f Dockerfile .
	docker tag $(IMAGE_BUILD_BRANCH) $(IMAGE_BUILD_COMMIT)
	# temporarily building latest here, too
	docker tag $(IMAGE_BUILD_COMMIT) $(IMAGE_BUILD_LATEST) 

.PHONY: build
build: clean ${RUNNER}

.PHONY: clean
clean:
	rm -Rf ${RUNNER}

.PHONY: image
image: $(IMAGE_BUILD_NAME)



	
	