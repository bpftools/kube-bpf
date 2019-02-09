#!/bin/bash

set -euo pipefail

DUMMY_SRC=dummy.c
DUMMY_OBJ=dummy.o

# obtain shared object - double pass: llvm, bpf
# clang -O2 -emit-llvm -c ${DUMMY_SRC} -o - | llc -march=bpf -filetype=obj -o ${DUMMY_OBJ}

# obtain shared object
clang -O2 -target bpf -c ${DUMMY_SRC} -o ${DUMMY_OBJ}

# obtain assembly
# clang -O2 -target bpf -c dummy.c -S