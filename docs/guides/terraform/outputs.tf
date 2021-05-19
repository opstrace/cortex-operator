
output "aws_region" {
  value = var.aws_region
}

output "cluster_name" {
  value = local.cluster_name
}

output "cortex_data_bucket_name" {
  value = local.cortex_data_bucket_name
}

output "cortex_config_bucket_name" {
  value = local.cortex_config_bucket_name
}
