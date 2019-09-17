# kube-bpf

> **WIP**

BPFs on Kubernetes.

![BPF custom resources](/docs/images/operator.png)

## Usage

1. Build the operator, the eBPF runner, and other Go tools

   ```console
   make
   ```

2. Build the docker image (and push it when the runner has changes)

    ```console
    make image
    docker push bpftools/runbpf
    ```

3. Create the BPF custom resources

    ```console
    make examples
    ```

    This command creates BPF custom resources - eg., YAML files - for the eBPF programs listend in `BPF_SOURCES` variable.
    In case you want to scope the resources you can issue the `make BPF_NAMESPACE=awesome examples` command.
    You can modify the `BPF_SOURCES` and `BPF_NAMES` variables appending your eBPF programs to make it compile also them.

4. Start the operator

    ```console
    ./output/operator
    ```

5. Apply the BPF, eg.:

    ```console
    kubectl apply -f output/pacchetti.yaml
    ```
