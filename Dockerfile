# FROM golang:1.11.5-alpine3.8 as gobuilder

# RUN apk add --update \
#     make \
#     bash \
#     git 
#     # elfutils-dev \
#     # musl-dev \
#     # gcc 

# RUN ln -s /usr/lib/cmake/llvm5 /usr/lib/cmake/llvm
# RUN ln -s /usr/include/llvm5/llvm /usr/include/llvm
# RUN ln -s /usr/include/llvm5/llvm-c /usr/include/llvm-c

# ADD . /go/src/github.com/bpftools/kube-bpf
# WORKDIR /go/src/github.com/bpftools/kube-bpf

# RUN make build

FROM alpine:3.8

RUN apk add --update libc6-compat

# RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
# COPY --from=gobuilder /go/src/github.com/bpftools/kube-bpf/output/runner /bin/bpfrun

# ENTRYPOINT ["/bin/bpfrun"]

# temporary
ADD output/runner /runner
ENTRYPOINT ["/runner"]

# run me => docker run -it --cap-add SYS_ADMIN -v /sys:/sys -p 9387:9387 bpftools/runbpf
