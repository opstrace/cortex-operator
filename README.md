# cortex-operator

The cortex-operator is a project to manage the lifecycle of Cortex in Kubernetes.

## Requirements

### Build

- Docker
- Kubectl

### Run
- Kubernetes cluster 1.18+

## Installation

Build and push your Docker image to the location specified by `IMG`:

```
make docker-build docker-push IMG=<registry/cortex-operator:canary>
```

Install the CRDs in the cluster:

```
make install
```

Deploy the controller to the cluster with the Docker image specified by `IMG`:

```
make deploy IMG=<registry/cortex-operator:canary>
```

## Uninstall

Remove the controller from the cluster:

```
make undeploy
```

Remove the CRDs from the cluster:

```
make uninstall
```

## Roadmap

- Deploy Cortex in different topologies
- Automate moving workloads to other instance
- Auto-scaling of services

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[Apache License, Version 2](http://www.apache.org/licenses/LICENSE-2.0)
