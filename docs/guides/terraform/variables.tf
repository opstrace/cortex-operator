variable "cluster_name_prefix" {
  type = string
  default = "cortex-operator-example"
}

variable "kubernetes_version" {
  type = string
  default = "1.18"
}

variable "aws_region" {
  type = string
  default = "us-west-2"
}

resource "random_id" "cluster_name" {
  byte_length = 2
  prefix = "${var.cluster_name_prefix}-"
}

locals {
  cluster_name = "${random_id.cluster_name.hex}"
  cortex_data_bucket_name = "${random_id.cluster_name.hex}-data"
  cortex_config_bucket_name = "${random_id.cluster_name.hex}-config"
}
