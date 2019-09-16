# kube-bpf

> **WIP**

BPFs on Kubernetes.

![BPF custom resources](/docs/images/operator.png)

## Usage

1. Create the BPF custom resources

    The makefile provides a `make examples` which creates BPF custom resources - eg., YAML files - for the existing eBPF in the example directory.
    In case you want to scope the resources you can issue the `make BPF_NAMESPACE=awesome examples` command.
    You can modify the `BPF_SOURCES` and `BPF_NAMES` variables appending your eBPF programs to make it compile also them.

**WIP**

