# ...

## Create a config map

Given an _object file_ we want to put it into a config map.

How to do this?

```bash
kubectl create configmap --from-file path/to/superpowers.o superpowersname -o yaml --dry-run
```

Just copy it.
