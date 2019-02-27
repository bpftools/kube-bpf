# kube-bpf

> **WIP**

BPFs on Kubernetes.

![BPF custom resources](/docs/images/operator.png)

## Create a config map

Given an _object file_ we want to put it into a config map.

How to do this?

```bash
kubectl create configmap --from-file path/to/superpowers.o superpowersname -n namespace -o yaml --dry-run
```

Or look at `examples/build.sh` for now.
