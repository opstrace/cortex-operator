# cortex-operator

The cortex-operator is a project to manage the lifecycle of [Cortex](https://cortexmetrics.io/) in Kubernetes.

**Project status: alpha** Not all planned features are completed. The API, spec, status and other user facing objects will, most likely, change. *Don't use it in production.*

## Requirements

### Build

- Docker
- Kubectl

### Run

- [EKS](https://aws.amazon.com/eks/) cluster version 1.18+
- Two S3 buckets
- AWS IAM policy for nodes in EKS cluster to access S3 buckets
- Kubectl configured to access the EKS cluster

Check this [guide](./docs/guides/terraform/README.md) on how to set up the infrastructure with Terraform.

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

## Quickstart

You can use this [guide](./docs/guides/terraform/README.md) to deploy the required infrastructure with Terraform.

Edit the [sample resource](./config/samples/cortex_v1alpha1_cortex.yaml) and set the bucket name for the blocks storage. If you used our guide, it'd be in the form of `cortex-operator-example-XXXX-data`. Set the bucket name for the alert manager and ruler configuration. If you used our guide, it'd be in the form of `cortex-operator-example-XXXX-config`

Create a Cortex resource to trigger the cortex-operator to start a deployment:

```
kubectl apply -f config/samples/cortex_v1alpha1_cortex.yaml
```

You should see a flurry of activity in the logs of the `cortex-operator`:

```
kubectl logs -n cortex-operator-system deploy/cortex-operator-controller-manager manager
```

You can confirm all the pods are up and running:

```
kubectl get pods
```

The output should be something like this:

```
NAME                             READY   STATUS    RESTARTS   AGE
compactor-0                      1/1     Running   0          2m59s
compactor-1                      1/1     Running   0          2m46s
distributor-db7f645c7-2bzlj      1/1     Running   0          2m59s
distributor-db7f645c7-2n22k      1/1     Running   0          2m59s
ingester-0                       1/1     Running   0          2m59s
ingester-1                       1/1     Running   0          1m26s
memcached-0                      1/1     Running   0          3m
memcached-index-queries-0        1/1     Running   0          3m
memcached-index-writes-0         1/1     Running   0          3m
memcached-metadata-0             1/1     Running   0          3m
memcached-results-0              1/1     Running   0          3m
querier-7dbd4cb465-66q95         1/1     Running   0          2m59s
querier-7dbd4cb465-frfnj         1/1     Running   1          2m59s
query-frontend-b9f7f97b7-g7lsf   1/1     Running   0          2m59s
query-frontend-b9f7f97b7-tsppd   1/1     Running   0          2m59s
store-gateway-0                  1/1     Running   0          2m59s
store-gateway-1                  1/1     Running   0          2m34s
```

You can now send metrics to Cortex. As an example, let's set up Grafana Agent to collect metrics from the Kubernetes nodes and send them to Cortex.

Create all the resources with:

```
kubectl apply -f docs/samples/grafana-example-manifest.yaml
```

In the future, we'll set up a Grafana dashboard to check these metrics, but for now, we'll use [cortex-tools](https://github.com/grafana/cortex-tools) to confirm Cortex is receiving metrics from the Grafana Agent.

Set up port-forward with kubectl to query Cortex:

```
kubectl port-forward svc/query-frontend 8080:80
```

In another terminal run
```
cortextool remote-read dump --address=http://localhost:8080 --remote-read-path=/api/v1/read
```

The output should be something like this:

```
INFO[0000] Created remote read client using endpoint 'http://localhost:8080/api/v1/read'
INFO[0000] Querying time from=2021-05-19T12:25:29Z to=2021-05-19T13:25:29Z with selector=up
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430648242
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430663242
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430678242
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430693242
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430708242
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430723242
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430635807
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430650807
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430665807
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430680807
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430695807
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430710807
{__name__="up", instance="ip-10-0-0-42.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430725810
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430641820
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430656820
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430671820
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430686820
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430701820
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/agent", namespace="monitoring"} 1 1621430716820
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430645938
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430660938
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430675938
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430690938
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430705938
{__name__="up", instance="ip-10-0-1-47.us-west-2.compute.internal", job="monitoring/node-exporter", namespace="monitoring"} 1 1621430720938
```

## Roadmap

- Webhook validation of Cortex configuration
- Deploy Cortex in different topologies
- Automate moving workloads to other instance
- Auto-scaling of services

## Contributing

Pull requests are welcome. For significant changes, please open an issue first to discuss what you would like to change.

## License
[Apache License, Version 2](http://www.apache.org/licenses/LICENSE-2.0)

