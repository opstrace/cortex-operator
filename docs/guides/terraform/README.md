# How to create an EKS cluster using Terraform and deploy cortex-operator

This guide shows how you can create an EKS cluster and the required S3 bucket with Terraform to deploy the cortex-operator.


## Requirements

- AWS account and credentials set up for your environment
- [Terraform](https://learn.hashicorp.com/tutorials/terraform/install-cli?in=terraform/aws-get-started)

## Acknowledgement

The Terraform files are based on the [Terraform EKS provider examples](https://github.com/hashicorp/terraform-provider-kubernetes/tree/main/_examples/eks).

## Preamble

The base directory to create the infrastructure must be `examples/terraform` in the repo root directory.

```
git clone https://github.com/opstrace/cortex-operator cortex-operator
cd cortex-operator/examples/terraform
```

Proceed to initialize terraform.

```
terraform init
```

## Configure

Edit `example.tfvars` with your favorite editor and choose the cluster name prefix, the AWS region, and Kubernetes cluster version.

The current defaults are:

```
cluster_name_prefix = "cortex-operator-example"
aws_region = "us-west-2"
kubernetes_version = "1.18"
```

## Create Infrastructure

To create all the required infrastructure to run the cortex-operator:

```
terraform apply -var-file=example.tfvars
```

Review the plan and apply it by typing `yes` when the following prompt appears:

```
Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes
```

It takes, on average, 15 minutes to create the EKS cluster and required S3 buckets. When it's done, it'll print out the cluster name and names of S3 buckets it created.

This is how it can look on a successful run when it's finished:

```
Outputs:

aws_region = "us-west-2"
cluster_name = "cortex-operator-example-209f"
cortex_config_bucket_name = "cortex-operator-example-209f-config"
cortex_data_bucket_name = "cortex-operator-example-209f-data"
```

## Set up kubeconfig

When you've provisioned your EKS cluster, you need to configure kubectl.

Run the following command to retrieve the access credentials for your cluster and automatically configure kubectl.

```
aws eks --region $(terraform output --raw aws_region) update-kubeconfig --name $(terraform output --raw cluster_name)
```

## Teardown Infrastructure

When you are done, don't forget to clean up everything with:

```
terraform destroy
```
