SHELL=/bin/bash -o pipefail

COMMIT_NUM := $(shell git rev-parse HEAD 2> /dev/null || true)
GIT_COMMIT := $(if $(shell git status --porcelain --untracked-files=no),${COMMIT_NUM}-dirty,${COMMIT_NUM})
GIT_BRANCH := $(shell echo $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null) | sed -e "s/[^[:alnum:]]/-/g")

IMAGE_BUILD_NAME ?= bpftools/runbpf
IMAGE_BUILD_FLAG ?= --no-cache

IMAGE_BUILD_BRANCH := ${IMAGE_BUILD_NAME}:${GIT_BRANCH}
IMAGE_BUILD_COMMIT := ${IMAGE_BUILD_NAME}:${GIT_COMMIT}
IMAGE_BUILD_LATEST := ${IMAGE_BUILD_NAME}:latest

LDFLAGS := -ldflags ""# -ldflags "-extldflags '-static'"

RUNNER := output/runner
OPERATOR := output/operator
GENYAML := output/genyaml

# List of Go programs to build
GO_PROGRAMS := ${RUNNER} ${OPERATOR} ${GENYAML}
GO_SOURCES := ./cmd/runner ./cmd/operator ./cmd/tools/genyaml

# List of BPF programs you want to build
BPF_SOURCES = examples/dummy.c examples/pkts.c
BPF_NAMES = dummy pacchetti

BPF_PROGRAMS_DIR := output
BPF_YAML_OUT := $(patsubst %, ${BPF_PROGRAMS_DIR}/%.yaml, ${BPF_NAMES})
BPF_OBJS_OUT := $(patsubst %, ${BPF_PROGRAMS_DIR}/%.o, ${BPF_NAMES})

.PHONY: build
build: clean ${GO_PROGRAMS}

.PHONY: image
image: ${IMAGE_BUILD_NAME}

.PHONY: examples
examples: ${BPF_YAML_OUT}

.PHONY: clean
clean:
	@rm -Rf ${GO_PROGRAMS}

$(IMAGE_BUILD_NAME): ${RUNNER}
	docker build \
		${IMAGE_BUILD_FLAG} \
		-t ${IMAGE_BUILD_BRANCH} \
		-f Dockerfile .
	docker tag $({MAGE_BUILD_BRANCH} ${IMAGE_BUILD_COMMIT}
	# temporarily building latest here, too
	docker tag ${IMAGE_BUILD_COMMIT} ${IMAGE_BUILD_LATEST}

# (1): The Go program being built
# (2): The Go source file
define gobuilds
$(1): $(2)/main.go
	GO11MODULE=on go build ${LDFLAGS} -o $(1) $(2)
endef
$(foreach GO_TARGET,${GO_PROGRAMS},\
	$(eval $(call gobuilds,${GO_TARGET},$(firstword ${GO_SOURCES})))\
	$(eval GO_SOURCES := $(wordlist 2,$(words ${GO_SOURCES}),${GO_SOURCES}))\
)

# (1): The BPF source file being built
# (2): The K8S name of the resource being built
# (3): The BPF object output file
# (4): The BPF YAML output file
define bpfbuilds
.PHONY: $(3)
$(3): $(1)
	@rm -f $(4)
	@echo -e "------\n[eBPF] $(1) => $(3)"
	@clang -O2 -target bpf -c $(1) -o $(3)
$(4): ${GENYAML} $(3)
ifeq ($(strip ${BPF_NAMESPACE}),)
	@echo -e "------\n[YAML] $(3) => $(4)"
	@${GENYAML} -r $(2) --ending-dashes >> $(4)
	@kubectl create configmap --from-file $(3) $(2)-config -o yaml --dry-run >> $(4)
else
	@echo -e "------\n[YAML] $(3) => $(4)"
	@${GENYAML} -r $(2) --ending-dashes -n ${BPF_NAMESPACE} >> $(4)
	@kubectl create configmap --from-file $(3) $(2)-config -o yaml --dry-run -n ${BPF_NAMESPACE} >> $(4)
endif
	@rm -f $(3)
endef
$(foreach BPF,${BPF_SOURCES},\
	$(eval $(call bpfbuilds,${BPF},$(firstword ${BPF_NAMES}),$(firstword ${BPF_OBJS_OUT}),$(firstword ${BPF_YAML_OUT})))\
	$(eval BPF_NAMES := $(wordlist 2,$(words ${BPF_NAMES}),${BPF_NAMES}))\
	$(eval BPF_OBJS_OUT := $(wordlist 2,$(words ${BPF_OBJS_OUT}),${BPF_OBJS_OUT}))\
	$(eval BPF_YAML_OUT := $(wordlist 2,$(words ${BPF_YAML_OUT}),${BPF_YAML_OUT}))\
)